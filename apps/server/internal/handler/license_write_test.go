package handler_test

// license_write_test.go — 许可写入端点集成测试。
// 经真实 JWT 角色（register 后 token）走 setupServer 的完整 HTTP 路径；
// 需要 TEST_DATABASE_URL_HANDLER，未设时 skip（如实，不假跑）。

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// 复用 handler_test 的 setupServer/doJSON（同包）。

func registerToken(t *testing.T, ts *httptest.Server, email string) string {
	t.Helper()
	code, body := doJSON(t, "POST", ts.URL+"/api/v1/auth/register", map[string]any{
		"company": "许可测试", "name": "测试", "email": email, "password": "password123",
	}, "")
	if code != http.StatusCreated {
		t.Fatalf("register %d body %v", code, body)
	}
	tok, _ := body.(map[string]any)["token"].(string)
	if tok == "" {
		t.Fatal("register returned no token")
	}
	return tok
}

func TestLicenseWriteIntegration(t *testing.T) {
	ts := setupServer(t)
	token := registerToken(t, ts, "license-write@example.com")

	// 未认证写入 → 401
	code, _ := doJSON(t, "POST", ts.URL+"/api/v1/license", map[string]any{
		"deployment_id": "d1", "claims": map[string]any{"licensee": "acme", "feature_grants": []any{}},
	}, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated write status %d, want 401", code)
	}

	// 认证 + 无效 claims → 400
	code, _ = doJSON(t, "POST", ts.URL+"/api/v1/license", map[string]any{
		"deployment_id": "d1", "claims": "not-json",
	}, token)
	if code != http.StatusBadRequest {
		t.Fatalf("invalid claims status %d, want 400", code)
	}

	// 认证 + 空 grants → 400
	code, _ = doJSON(t, "POST", ts.URL+"/api/v1/license", map[string]any{
		"deployment_id": "d1", "claims": map[string]any{"licensee": "acme", "feature_grants": []any{}},
	}, token)
	if code != http.StatusBadRequest {
		t.Fatalf("empty grants status %d, want 400", code)
	}

	// 认证 + 合法 claims → 200
	code, _ = doJSON(t, "POST", ts.URL+"/api/v1/license", map[string]any{
		"deployment_id": "d-write",
		"claims": map[string]any{
			"licensee": "acme", "license_revision": 5,
			"feature_grants": []any{
				map[string]any{"grant_id": "g1", "feature_key": "archive.read", "state": "active", "valid_from": "2026-01-01T00:00:00Z"},
			},
		},
	}, token)
	if code != http.StatusOK {
		t.Fatalf("valid write status %d, want 200", code)
	}
}
