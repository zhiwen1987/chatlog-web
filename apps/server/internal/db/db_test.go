package db

// db_test.go — license claims 持久化集成测试。
// 需要 TEST_DATABASE_URL 指向可用 PostgreSQL；未设时 skip（如实，不假跑）。

import (
	"context"
	"os"
	"testing"
	"time"

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
