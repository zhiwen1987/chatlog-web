package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/db"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/handler"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/migration"
)

func setupServer(t *testing.T) *httptest.Server {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL_HANDLER")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL_HANDLER not set; skipping integration test")
	}
	conn, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := migration.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	srv := handler.New(handler.Deps{
		DB:          conn,
		JWTSecret:   "test-secret",
		TokenTTLMin: 60,
	})
	ts := httptest.NewServer(srv.Routes())
	t.Cleanup(ts.Close)
	return ts
}

func doJSON(t *testing.T, method, url string, body any, token string) (int, any) {
	t.Helper()
	var rdr *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	} else {
		rdr = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	var out any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode %s: %v", url, err)
	}
	return resp.StatusCode, out
}

func TestRegisterLoginMe(t *testing.T) {
	ts := setupServer(t)

	email := fmt.Sprintf("reg-%d@example.com", time.Now().UnixNano())
	// 注册
	code, body := doJSON(t, "POST", ts.URL+"/api/v1/auth/register", map[string]any{
		"company": "测试公司", "name": "张三", "email": email, "password": "password123",
	}, "")
	if code != http.StatusCreated {
		t.Fatalf("register status %d body %v", code, body)
	}
	token, _ := body.(map[string]any)["token"].(string)
	if token == "" {
		t.Fatal("register returned no token")
	}

	// 未认证访问 me -> 401
	code, _ = doJSON(t, "GET", ts.URL+"/api/v1/auth/me", nil, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated me status %d, want 401", code)
	}

	// me
	code, body = doJSON(t, "GET", ts.URL+"/api/v1/auth/me", nil, token)
	if code != http.StatusOK {
		t.Fatalf("me status %d body %v", code, body)
	}
	if body.(map[string]any)["role"] != "owner" {
		t.Fatalf("me role %v, want owner", body)
	}

	// 重复注册同邮箱 -> 409
	code, _ = doJSON(t, "POST", ts.URL+"/api/v1/auth/register", map[string]any{
		"company": "B", "name": "李四", "email": email, "password": "password123",
	}, "")
	if code != http.StatusConflict {
		t.Fatalf("duplicate register status %d, want 409", code)
	}

	// 登录
	code, body = doJSON(t, "POST", ts.URL+"/api/v1/auth/login", map[string]any{
		"email": email, "password": "password123",
	}, "")
	if code != http.StatusOK {
		t.Fatalf("login status %d body %v", code, body)
	}
	if body.(map[string]any)["role"] != "owner" {
		t.Fatalf("login role %v, want owner", body)
	}

	// 错误密码 -> 401
	code, _ = doJSON(t, "POST", ts.URL+"/api/v1/auth/login", map[string]any{
		"email": email, "password": "wrong",
	}, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("bad login status %d, want 401", code)
	}
}

func TestAdminLists(t *testing.T) {
	ts := setupServer(t)
	_, regBody := doJSON(t, "POST", ts.URL+"/api/v1/auth/register", map[string]any{
		"company": "A", "name": "王五", "email": fmt.Sprintf("adm-%d@example.com", time.Now().UnixNano()), "password": "password123",
	}, "")
	token, _ := regBody.(map[string]any)["token"].(string)

	// tenants
	code, body := doJSON(t, "GET", ts.URL+"/api/v1/tenants", nil, token)
	if code != http.StatusOK {
		t.Fatalf("tenants status %d body %v", code, body)
	}
	// users
	code, body = doJSON(t, "GET", ts.URL+"/api/v1/users", nil, token)
	if code != http.StatusOK {
		t.Fatalf("users status %d body %v", code, body)
	}
	if !strings.Contains(mustJSON(t, body), "王五") {
		t.Fatalf("users should contain 王五: %v", body)
	}
	// devices
	code, _ = doJSON(t, "GET", ts.URL+"/api/v1/devices", nil, token)
	if code != http.StatusOK {
		t.Fatalf("devices status %d", code)
	}
	// sources
	code, _ = doJSON(t, "GET", ts.URL+"/api/v1/sources", nil, token)
	if code != http.StatusOK {
		t.Fatalf("sources status %d", code)
	}
	// audit
	code, _ = doJSON(t, "GET", ts.URL+"/api/v1/audit", nil, token)
	if code != http.StatusOK {
		t.Fatalf("audit status %d", code)
	}
}

func TestHealth(t *testing.T) {
	ts := setupServer(t)
	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("healthz status %d", resp.StatusCode)
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}
