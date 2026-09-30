package handler_test

// device_test.go — 设备注册端点集成测试（R42.7 设备登记）。
// 覆盖：注册成功、平台枚举拒绝、认证缺失 401、同 fingerprint 幂等（返回同 id）。

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// registerDevice 注册设备并返回 device_id（复用 registerToken 拿 token）。
func registerDevice(t *testing.T, ts *httptest.Server, token, fp string) (int, map[string]any) {
	t.Helper()
	code, body := doJSON(t, "POST", ts.URL+"/api/v1/devices/register", map[string]any{
		"device_name": "办公机", "device_fingerprint": fp,
		"platform": "macos", "architecture": "arm64", "client_version": "1.0.0",
	}, token)
	m, _ := body.(map[string]any)
	return code, m
}

func TestRegisterDevice(t *testing.T) {
	ts := setupServer(t)
	token := registerToken(t, ts, "device@example.com")
	fp := fmt.Sprintf("fp-%d", time.Now().UnixNano())

	code, m := registerDevice(t, ts, token, fp)
	if code != http.StatusOK {
		t.Fatalf("register device: got %d body %v", code, m)
	}
	if m["status"] != "active" || m["device_id"] == "" {
		t.Fatalf("register device unexpected body: %v", m)
	}
	firstID := m["device_id"].(string)

	// 同 fingerprint 幂等：二次注册返回同 id
	code, m = registerDevice(t, ts, token, fp)
	if code != http.StatusOK {
		t.Fatalf("re-register device: got %d", code)
	}
	if id := m["device_id"].(string); id != firstID {
		t.Fatalf("idempotent re-register: got %s want %s", id, firstID)
	}
}

func TestRegisterDeviceRejectsBadPlatform(t *testing.T) {
	ts := setupServer(t)
	token := registerToken(t, ts, "device.bad@example.com")

	code, _ := doJSON(t, "POST", ts.URL+"/api/v1/devices/register", map[string]any{
		"device_name": "x", "device_fingerprint": "fp-bad",
		"platform": "ios", "architecture": "arm64",
	}, token)
	if code != http.StatusBadRequest {
		t.Fatalf("bad platform: got %d want 400", code)
	}
}

func TestRegisterDeviceRequiresAuth(t *testing.T) {
	ts := setupServer(t)
	code, _ := doJSON(t, "POST", ts.URL+"/api/v1/devices/register", map[string]any{
		"device_name": "x", "device_fingerprint": "fp-unauth",
		"platform": "macos", "architecture": "arm64",
	}, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("unauth register device: got %d want 401", code)
	}
}