package handler

// license_test.go — 许可状态端点测试。

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

func licenseServer(claims *model.Claims) *Server {
	return New(Deps{JWTSecret: "test-secret", TokenTTLMin: 60, License: claims})
}

func TestLicenseStatusNilClaims(t *testing.T) {
	srv := licenseServer(nil)
	req := httptest.NewRequest("GET", "/api/v1/license/status", nil)
	rec := httptest.NewRecorder()
	srv.licenseStatus(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if body["present"] != false {
		t.Fatalf("expected present=false for nil claims, got %v", body["present"])
	}
	feat := body["features"].(map[string]any)
	ar := feat["archive.read"].(map[string]any)
	if ar["Allowed"] != false {
		t.Fatalf("expected archive.read denied for nil claims")
	}
}

func TestLicenseStatusEmptyClaims(t *testing.T) {
	srv := licenseServer(&model.Claims{})
	req := httptest.NewRequest("GET", "/api/v1/license/status", nil)
	rec := httptest.NewRecorder()
	srv.licenseStatus(rec, req)
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	feat := body["features"].(map[string]any)
	ar := feat["archive.read"].(map[string]any)
	if ar["Allowed"] != false {
		t.Fatalf("expected archive.read denied for empty claims")
	}
}

func TestLicenseStatusGranted(t *testing.T) {
	c := &model.Claims{
		Licensee:   "acme",
		Deployment: "d1",
		FeatureGrants: []model.Grant{
			{GrantID: "g1", FeatureKey: "archive.read", State: "active",
				ValidFrom: "2026-01-01T00:00:00Z", ValidUntil: "2027-01-01T00:00:00Z"},
			{GrantID: "g2", FeatureKey: "archive.ingest", State: "active",
				ValidFrom: "2026-01-01T00:00:00Z", ValidUntil: "2027-01-01T00:00:00Z"},
		},
	}
	srv := licenseServer(c)
	req := httptest.NewRequest("GET", "/api/v1/license/status", nil)
	rec := httptest.NewRecorder()
	srv.licenseStatus(rec, req)
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["present"] != true {
		t.Fatalf("expected present=true")
	}
	feat := body["features"].(map[string]any)
	ar := feat["archive.read"].(map[string]any)
	if ar["Allowed"] != true {
		t.Fatalf("expected archive.read allowed")
	}
	mi := feat["media.upload.image"].(map[string]any)
	if mi["Allowed"] != false {
		t.Fatalf("expected media.upload.image denied (missing archive.ingest dep)")
	}
}