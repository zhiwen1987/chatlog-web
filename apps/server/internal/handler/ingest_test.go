package handler_test

// ingest_test.go — 数据链接入端点集成测试（R42 ingestion）。
// 覆盖：archive.ingest 未授权 403；授权后接入 accepted=1 幂等重发 duplicated=1；
// 未认证 401；枚举非法 400。真实 PG + 真实迁移（006 唯一索引）。

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

func ingestLicense() *model.Claims {
	return &model.Claims{
		Licensee: "acme",
		FeatureGrants: []model.Grant{{
			GrantID: "g-ingest", FeatureKey: "archive.ingest",
			State: "active", ValidFrom: "2026-01-01T00:00:00Z", ValidUntil: "2027-01-01T00:00:00Z",
		}},
	}
}

func ingestBody(prefix string) map[string]any {
	return map[string]any{
		"upstream_message_id": "up-" + prefix, "source_type": "windows_wechat",
		"source_external_id": "wx-" + prefix, "source_display_name": "接入号",
		"conversation_ref": "conv-" + prefix, "conversation_type": "private",
		"direction": "incoming", "message_type": "text",
		"sent_at": "2026-01-01T00:00:00Z", "content_text": "hi", "decode_status": "ok",
	}
}

func TestIngestRequiresFeature(t *testing.T) {
	ts := setupServer(t) // 无 License → archive.ingest 默认拒绝
	token := registerToken(t, ts, "ingest.deny@example.com")
	code, body := doJSON(t, "POST", ts.URL+"/api/v1/ingest/messages", ingestBody("deny"), token)
	if code != http.StatusForbidden {
		t.Fatalf("ingest without feature: got %d body %v want 403", code, body)
	}
}

func TestIngestAcceptedAndIdempotent(t *testing.T) {
	ts := setupServer(t, ingestLicense())
	token := registerToken(t, ts, "ingest.ok@example.com")
	prefix := fmt.Sprintf("ok-%d", time.Now().UnixNano())

	code, body := doJSON(t, "POST", ts.URL+"/api/v1/ingest/messages", ingestBody(prefix), token)
	if code != http.StatusOK {
		t.Fatalf("ingest: got %d body %v", code, body)
	}
	m, _ := body.(map[string]any)
	if m["accepted"].(float64) != 1 || m["duplicated"].(float64) != 0 {
		t.Fatalf("ingest accepted: %v want accepted=1 dup=0", body)
	}

	// 幂等重发：同 upstream_message_id → duplicated=1, accepted=0
	code, body = doJSON(t, "POST", ts.URL+"/api/v1/ingest/messages", ingestBody(prefix), token)
	if code != http.StatusOK {
		t.Fatalf("re-ingest: got %d body %v", code, body)
	}
	m, _ = body.(map[string]any)
	if m["accepted"].(float64) != 0 || m["duplicated"].(float64) != 1 {
		t.Fatalf("re-ingest: %v want accepted=0 dup=1", body)
	}
}

func TestIngestRequiresAuth(t *testing.T) {
	ts := setupServer(t, ingestLicense())
	code, _ := doJSON(t, "POST", ts.URL+"/api/v1/ingest/messages", ingestBody("unauth"), "")
	if code != http.StatusUnauthorized {
		t.Fatalf("unauth ingest: got %d want 401", code)
	}
}

func TestIngestBadEnum(t *testing.T) {
	ts := setupServer(t, ingestLicense())
	token := registerToken(t, ts, "ingest.bad@example.com")
	b := ingestBody("bad")
	b["source_type"] = "bogus"
	code, _ := doJSON(t, "POST", ts.URL+"/api/v1/ingest/messages", b, token)
	if code != http.StatusBadRequest {
		t.Fatalf("bad source_type: got %d want 400", code)
	}
}