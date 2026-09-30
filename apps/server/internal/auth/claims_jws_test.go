package auth

// claims_jws_test.go — JWS 签发/验签闭环（R42.8）。
// 全部自生成测试密钥（HS256 对称 + RS256 非对称），不依赖真实签发方。
// 覆盖：正签正验、篡改拒绝、错钥拒绝、过期拒绝、aud 不匹配拒绝。

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

func testPayload() *ClaimsJWSPayload {
	return &ClaimsJWSPayload{
		Typ:               "license-claims.v2",
		Licensee:          "acme",
		Deployment:        "deploy-1",
		FeatureCatalogVer: 1,
		LicenseRevision:   1,
		FeatureGrants: []model.Grant{
			{GrantID: "g1", FeatureKey: "chatlog", State: "active", ValidFrom: time.Now().Add(-time.Hour).Format(time.RFC3339), ValidUntil: time.Now().Add(24 * time.Hour).Format(time.RFC3339)},
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "issuer.test",
			Audience:  jwt.ClaimStrings{"deploy-1"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
}

func TestSignVerifyHS256(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // 32B
	tok, err := SignClaims(key, testPayload())
	if err != nil {
		t.Fatalf("sign hs256: %v", err)
	}
	claims, err := VerifyClaimsJWS(key, tok, "deploy-1")
	if err != nil {
		t.Fatalf("verify hs256: %v", err)
	}
	if claims.Licensee != "acme" || claims.LicenseRevision != 1 || len(claims.FeatureGrants) != 1 {
		t.Fatalf("verify hs256 unexpected claims: %+v", claims)
	}
}

func TestVerifyRejectsTampered(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	tok, err := SignClaims(key, testPayload())
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	tampered := tok[:len(tok)-4] + "AAAA"
	if _, err := VerifyClaimsJWS(key, tampered, "deploy-1"); err != ErrLicenseSignature {
		t.Fatalf("tampered: got %v want ErrLicenseSignature", err)
	}
}

func TestVerifyRejectsWrongKey(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	other := []byte("fedcba9876543210fedcba9876543210")
	tok, _ := SignClaims(key, testPayload())
	if _, err := VerifyClaimsJWS(other, tok, "deploy-1"); err != ErrLicenseSignature {
		t.Fatalf("wrong key: got %v want ErrLicenseSignature", err)
	}
}

func TestVerifyRejectsExpired(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	p := testPayload()
	p.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))
	tok, _ := SignClaims(key, p)
	if _, err := VerifyClaimsJWS(key, tok, "deploy-1"); err != ErrLicenseExpired {
		t.Fatalf("expired: got %v want ErrLicenseExpired", err)
	}
}

func TestVerifyRejectsWrongAudience(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	tok, _ := SignClaims(key, testPayload())
	if _, err := VerifyClaimsJWS(key, tok, "other-deploy"); err != ErrLicenseSignature {
		t.Fatalf("wrong aud: got %v want ErrLicenseSignature", err)
	}
}

func TestSignVerifyRS256(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("gen rsa: %v", err)
	}
	tok, err := SignClaims(priv, testPayload())
	if err != nil {
		t.Fatalf("sign rs256: %v", err)
	}
	claims, err := VerifyClaimsJWS(&priv.PublicKey, tok, "deploy-1")
	if err != nil {
		t.Fatalf("verify rs256: %v", err)
	}
	if claims.Licensee != "acme" {
		t.Fatalf("verify rs256 unexpected: %+v", claims)
	}
}

func TestVerifyRejectsRSAWrongKey(t *testing.T) {
	priv1, _ := rsa.GenerateKey(rand.Reader, 2048)
	priv2, _ := rsa.GenerateKey(rand.Reader, 2048)
	tok, _ := SignClaims(priv1, testPayload())
	if _, err := VerifyClaimsJWS(&priv2.PublicKey, tok, "deploy-1"); err != ErrLicenseSignature {
		t.Fatalf("rsa wrong key: got %v want ErrLicenseSignature", err)
	}
}