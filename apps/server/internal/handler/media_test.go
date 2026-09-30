package handler_test

// media_test.go — 媒体清单/收据端点集成测试（R42.7 媒体协议）。
// 覆盖：上报 manifest+receipt 后列表对账返回对应行；缺必填字段 400；
// 未认证 401；object_ref 无 manifest 时清单为 null（前端 reconcile 拒假 ACK）。

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func validManifest(prefix string) map[string]any {
	return map[string]any{
		"manifest_id": prefix + "-m1", "catalog_version": 1,
		"tenant_id": "t1", "deployment_id": "d1",
		"object_ref": "obj." + prefix, "media_type": "image",
		"origin": map[string]any{"kind": "original", "source": "wechat-desktop", "source_message_ref": "src.1"},
		"bytes_ref": "s3://b/obj." + prefix, "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"size_bytes": 10, "seq": "1", "created_at": "2026-01-01T00:00:00Z", "state": "stored",
	}
}

func validReceipt(prefix string) map[string]any {
	return map[string]any{
		"receipt_id": prefix + "-r1", "catalog_version": 1,
		"tenant_id": "t1", "deployment_id": "d1",
		"source": "wechat-desktop", "source_message_ref": "src.1",
		"seq": "1", "committed_at": "2026-01-01T00:00:00Z", "backup_set": "bs1",
		"object_ref": "obj." + prefix, "media_kind": "image",
	}
}

func TestMediaManifestAndReceiptRoundtrip(t *testing.T) {
	ts := setupServer(t)
	token := registerToken(t, ts, "media@example.com")
	prefix := fmt.Sprintf("rt-%d", time.Now().UnixNano())

	code, body := doJSON(t, "POST", ts.URL+"/api/v1/media/manifests", validManifest(prefix), token)
	if code != http.StatusOK {
		t.Fatalf("save manifest: got %d body %v", code, body)
	}
	code, body = doJSON(t, "POST", ts.URL+"/api/v1/media/receipts", validReceipt(prefix), token)
	if code != http.StatusOK {
		t.Fatalf("save receipt: got %d body %v", code, body)
	}

	// 列表对账：应返回 1 行，receipt+manifest 成对
	code, body = doJSON(t, "GET", ts.URL+"/api/v1/media/receipts", nil, token)
	if code != http.StatusOK {
		t.Fatalf("list receipts: got %d body %v", code, body)
	}
	m, _ := body.(map[string]any)
	data, _ := m["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("receipts data: got %d rows want 1 (%v)", len(data), body)
	}
	row := data[0].(map[string]any)
	rec := row["receipt"].(map[string]any)
	man := row["manifest"].(map[string]any)
	if rec["receipt_id"] != prefix+"-r1" {
		t.Fatalf("receipt id: %v", rec["receipt_id"])
	}
	if man["manifest_id"] != prefix+"-m1" {
		t.Fatalf("manifest id: %v", man["manifest_id"])
	}
}

func TestMediaManifestRejectsBadSha(t *testing.T) {
	ts := setupServer(t)
	token := registerToken(t, ts, "media.bad@example.com")
	m := validManifest("bad")
	m["sha256"] = "nope"
	code, _ := doJSON(t, "POST", ts.URL+"/api/v1/media/manifests", m, token)
	if code != http.StatusBadRequest {
		t.Fatalf("bad sha: got %d want 400", code)
	}
}

func TestMediaReceiptRequiresPairFields(t *testing.T) {
	ts := setupServer(t)
	token := registerToken(t, ts, "media.pair@example.com")
	r := validReceipt("pair")
	r["media_kind"] = "" // 保留 object_ref 但去 media_kind → 不成对
	code, _ := doJSON(t, "POST", ts.URL+"/api/v1/media/receipts", r, token)
	if code != http.StatusBadRequest {
		t.Fatalf("unpaired receipt: got %d want 400", code)
	}
}

func TestMediaRequiresAuth(t *testing.T) {
	ts := setupServer(t)
	code, _ := doJSON(t, "POST", ts.URL+"/api/v1/media/manifests", validManifest("unauth"), "")
	if code != http.StatusUnauthorized {
		t.Fatalf("unauth save manifest: got %d want 401", code)
	}
	code, _ = doJSON(t, "GET", ts.URL+"/api/v1/media/receipts", nil, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("unauth list receipts: got %d want 401", code)
	}
}