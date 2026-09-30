package handler_test

// integrity_test.go — 消息完整性报告端点集成测试（R42.10）。
// 覆盖：真实 PG 插入三类消息后，/api/v1/integrity/report 计数正确；
// 未认证 401；无数据租户返回零计数而非伪造。

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/db"
)

// seedIntegrityMessages 直接插库造消息：verified(ok+hash)、pending(ok+空hash)、excluded(encrypted)。
// 返回 (tenant_id, source_account_id, conversation_id)。
func seedIntegrityMessages(t *testing.T, tenantID string) (string, string) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL_HANDLER")
	conn, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	ctx := context.Background()

	var srcID, convID string
	if err := conn.QueryRowContext(ctx, `
		INSERT INTO source_accounts (tenant_id, source_type, external_account_id, display_name)
		VALUES ($1, 'windows_wechat', $2, '测试号')
		RETURNING id`, tenantID, fmt.Sprintf("ext-%d", time.Now().UnixNano())).Scan(&srcID); err != nil {
		t.Fatalf("seed source_account: %v", err)
	}
	if err := conn.QueryRowContext(ctx, `
		INSERT INTO conversations (tenant_id, source_account_id, external_conversation_id, conversation_type)
		VALUES ($1, $2, $3, 'private')
		RETURNING id`, tenantID, srcID, fmt.Sprintf("conv-%d", time.Now().UnixNano())).Scan(&convID); err != nil {
		t.Fatalf("seed conversation: %v", err)
	}

	// verified: decode_status=ok + content_hash 非空
	// pending: decode_status=ok + content_hash 空
	// excluded: decode_status=encrypted
	inserts := []struct{ status, hash string }{
		{"ok", "a" + fmt.Sprint(time.Now().UnixNano())},
		{"ok", ""},
		{"encrypted", ""},
	}
	for _, ins := range inserts {
		if _, err := conn.ExecContext(ctx, `
			INSERT INTO messages (tenant_id, source_account_id, conversation_id, direction, message_type, sent_at, content_text, content_hash, decode_status)
			VALUES ($1, $2, $3, 'incoming', 'text', now(), 'x', $4, $5)`,
			tenantID, srcID, convID, nullable(ins.hash), ins.status); err != nil {
			t.Fatalf("seed message: %v", err)
		}
	}
	return srcID, convID
}

// nullable 空串转 NULL（content_hash 为空视为未校验）。
func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func TestIntegrityReportCounts(t *testing.T) {
	ts := setupServer(t)
	token := registerToken(t, ts, "integrity@example.com")

	// 拿 tenant_id
	code, body := doJSON(t, "GET", ts.URL+"/api/v1/tenants", nil, token)
	if code != http.StatusOK {
		t.Fatalf("tenants status %d", code)
	}
	raw := mustJSON(t, body)
	var list []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(raw), &list); err != nil || len(list) == 0 {
		t.Fatalf("tenants parse: %v raw=%s", err, raw)
	}
	tenantID := list[0].ID

	_, _ = seedIntegrityMessages(t, tenantID)

	code, body = doJSON(t, "GET", ts.URL+"/api/v1/integrity/report", nil, token)
	if code != http.StatusOK {
		t.Fatalf("integrity report: got %d body %v", code, body)
	}
	m, _ := body.(map[string]any)
	counts, _ := m["counts"].(map[string]any)
	if counts == nil {
		t.Fatalf("integrity report missing counts: %v", body)
	}
	if c := counts["verified"].(float64); c != 1 {
		t.Fatalf("verified: got %v want 1", counts["verified"])
	}
	if c := counts["pending"].(float64); c != 1 {
		t.Fatalf("pending: got %v want 1", counts["pending"])
	}
	if c := counts["excluded"].(float64); c != 1 {
		t.Fatalf("excluded: got %v want 1", counts["excluded"])
	}
	if c := counts["in_scope"].(float64); c != 2 {
		t.Fatalf("in_scope: got %v want 2", counts["in_scope"])
	}
	if c := counts["total_discovered"].(float64); c != 3 {
		t.Fatalf("total_discovered: got %v want 3", counts["total_discovered"])
	}
	if c := counts["source_missing"].(float64); c != 0 {
		t.Fatalf("source_missing: got %v want 0", counts["source_missing"])
	}
}

func TestIntegrityReportRequiresAuth(t *testing.T) {
	ts := setupServer(t)
	code, _ := doJSON(t, "GET", ts.URL+"/api/v1/integrity/report", nil, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("unauth integrity report: got %d want 401", code)
	}
}

func TestIntegrityReportEmptyTenant(t *testing.T) {
	ts := setupServer(t)
	token := registerToken(t, ts, "integrity.empty@example.com")
	code, body := doJSON(t, "GET", ts.URL+"/api/v1/integrity/report", nil, token)
	if code != http.StatusOK {
		t.Fatalf("empty tenant report: got %d body %v", code, body)
	}
	m, _ := body.(map[string]any)
	counts, _ := m["counts"].(map[string]any)
	if counts["total_discovered"].(float64) != 0 || counts["in_scope"].(float64) != 0 {
		t.Fatalf("empty tenant counts: %v", counts)
	}
}