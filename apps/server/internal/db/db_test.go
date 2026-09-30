package db

// db_test.go — license claims 持久化集成测试。
// 需要 TEST_DATABASE_URL 指向可用 PostgreSQL；未设时 skip（如实，不假跑）。

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/migration"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

const testDeployment = "deployment-integration-test"

func TestLoadLicenseClaimsIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := migration.Migrate(ctx, db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// 清掉该部署的旧行，保证测试幂等
	if _, err := db.ExecContext(ctx, `DELETE FROM license_claims WHERE deployment_id = $1`, testDeployment); err != nil {
		t.Fatalf("cleanup: %v", err)
	}

	claims := `{"licensee":"acme","deployment":"` + testDeployment + `","license_revision":3,` +
		`"feature_grants":[{"grant_id":"g1","feature_key":"archive.read","state":"active",` +
		`"valid_from":"2026-01-01T00:00:00Z","valid_until":"2027-01-01T00:00:00Z"}]}`
	if _, err := db.ExecContext(ctx, `
		INSERT INTO license_claims (licensee, deployment_id, claims_json, license_revision)
		VALUES ($1, $2, $3::jsonb, 3)`, "acme", testDeployment, claims); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got, err := LoadLicenseClaims(ctx, db, testDeployment)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got == nil {
		t.Fatal("expected claims, got nil")
	}
	if got.Licensee != "acme" {
		t.Fatalf("licensee = %q, want acme", got.Licensee)
	}
	if len(got.FeatureGrants) != 1 || got.FeatureGrants[0].FeatureKey != "archive.read" {
		t.Fatalf("unexpected grants: %+v", got.FeatureGrants)
	}
	if !got.HasFeature("archive.read", time.Now().UTC()) {
		t.Fatal("expected archive.read granted")
	}

	// 未找到的部署返回 nil（默认拒绝）
	missing, err := LoadLicenseClaims(ctx, db, "deployment-not-exist")
	if err != nil {
		t.Fatalf("load missing: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing deployment, got %+v", missing)
	}
}

var _ = model.GrantActive // 引用 model 常量确保包加载

// testVerifyKey 固定 32 字节 HS256 对称密钥（仅测试）。
var testVerifyKey = []byte("01234567890123456789012345678901")

// TestLoadLicenseClaimsVerifiedIntegration 验证 JWS 验签加载路径：
// 签发→入库→验签通过；错钥→拒绝；篡改→拒绝。
func TestLoadLicenseClaimsVerifiedIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := migration.Migrate(ctx, db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	deployment := testDeployment + "-verified"
	if _, err := db.ExecContext(ctx, `DELETE FROM license_claims WHERE deployment_id = $1`, deployment); err != nil {
		t.Fatalf("cleanup: %v", err)
	}

	signed, err := auth.SignClaims(testVerifyKey, &auth.ClaimsJWSPayload{
		Licensee:          "acme",
		Deployment:        deployment,
		FeatureCatalogVer: 1,
		LicenseRevision:   5,
		FeatureGrants: []model.Grant{{
			GrantID: "g1", FeatureKey: "archive.read", State: model.GrantActive,
			ValidFrom: "2026-01-01T00:00:00Z", ValidUntil: "2027-01-01T00:00:00Z",
		}},
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{deployment},
			Issuer:    "issuer.test",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO license_claims (licensee, deployment_id, claims_json, claims_jws, license_revision)
		VALUES ($1, $2, $3::jsonb, $4, 5)`, "acme", deployment, `{"licensee":"acme","deployment":"`+deployment+`","license_revision":5}`, signed); err != nil {
		t.Fatalf("insert: %v", err)
	}

	// 正确密钥：验签通过
	got, err := LoadLicenseClaimsVerified(ctx, db, deployment, testVerifyKey, deployment)
	if err != nil {
		t.Fatalf("load verified: %v", err)
	}
	if got == nil {
		t.Fatal("expected claims, got nil")
	}
	if got.Licensee != "acme" || got.LicenseRevision != 5 {
		t.Fatalf("unexpected claims: %+v", got)
	}
	if !got.HasFeature("archive.read", time.Now().UTC()) {
		t.Fatal("expected archive.read granted")
	}

	// 错钥：拒绝（ErrLicenseSignature 包装）
	wrongKey := []byte("fedcba9876543210fedcba9876543210")
	if _, err := LoadLicenseClaimsVerified(ctx, db, deployment, wrongKey, deployment); err == nil {
		t.Fatal("expected error with wrong key, got nil")
	} else if !strings.Contains(err.Error(), "signature") {
		t.Fatalf("expected signature error, got: %v", err)
	}

	// 篡改 claims_jws：拒绝（写入篡改后的 JWS，更高 revision 使其排最前）
	// 篡改签名段：拒绝（改 JWS 紧凑串最后一段 signature 的末字符，破坏签名）
	lastDot := strings.LastIndex(signed, ".")
	if lastDot < 0 || lastDot == len(signed)-1 {
		t.Fatalf("unexpected signed jws: %q", signed)
	}
	tampered := signed[:lastDot+1]
	if signed[len(signed)-1] == 'A' {
		tampered += "B"
	} else {
		tampered += "A"
	}
	if _, err := db.ExecContext(ctx, `
		UPDATE license_claims SET claims_json = $3::jsonb, claims_jws = $4, license_revision = 99, updated_at = now()
		WHERE deployment_id = $1 AND licensee = $2`,
		deployment, "acme", `{"licensee":"acme","deployment":"`+deployment+`","license_revision":99}`, tampered); err != nil {
		t.Fatalf("update tampered: %v", err)
	}
	if _, err := LoadLicenseClaimsVerified(ctx, db, deployment, testVerifyKey, deployment); err == nil {
		t.Fatal("expected error with tampered claims, got nil")
	}

	// 未找到部署：nil（默认拒绝）
	missing, err := LoadLicenseClaimsVerified(ctx, db, "deployment-not-exist", testVerifyKey, "aud")
	if err != nil {
		t.Fatalf("load missing: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing deployment, got %+v", missing)
	}
}
