package handler_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/db"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/handler"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/migration"
)

// seedEnterpriseData 注册租户并插入联系人/会话/消息，返回 token、测试服务器与会话 ID。
func seedEnterpriseData(t *testing.T) (string, *httptest.Server, string) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL_HANDLER")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL_HANDLER not set")
	}
	conn, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := migration.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	srv := handler.New(handler.Deps{DB: conn, JWTSecret: "test-secret", TokenTTLMin: 60})
	ts := httptest.NewServer(srv.Routes())
	t.Cleanup(ts.Close)

	// 每个测试独立邮箱，避免并行 409
	email := fmt.Sprintf("ent-%d@example.com", time.Now().UnixNano())
	code, body := doJSON(t, "POST", ts.URL+"/api/v1/auth/register", map[string]any{
		"company": "Ent", "name": "测试员", "email": email, "password": "password123",
	}, "")
	if code != http.StatusCreated {
		t.Fatalf("register %d body %v", code, body)
	}
	token := body.(map[string]any)["token"].(string)
	tenantID := body.(map[string]any)["tenant_id"].(string)

	ctx := context.Background()
	var acctID string
	if err := conn.QueryRowContext(ctx, `
		INSERT INTO source_accounts(tenant_id, source_type, external_account_id, display_name)
		VALUES ($1,'chatlog_http','demo-acc','验收账号') RETURNING id`, tenantID).Scan(&acctID); err != nil {
		t.Fatalf("source account: %v", err)
	}
	var contactID string
	if err := conn.QueryRowContext(ctx, `
		INSERT INTO contacts(tenant_id, source_account_id, external_contact_id, nickname, remark)
		VALUES ($1,$2,'wxid_demo','演示联系人','重要客户') RETURNING id`, tenantID, acctID).Scan(&contactID); err != nil {
		t.Fatalf("contact: %v", err)
	}
	var convID string
	if err := conn.QueryRowContext(ctx, `
		INSERT INTO conversations(tenant_id, source_account_id, external_conversation_id, conversation_type, title, last_message_at)
		VALUES ($1,$2,'conv-1','private','与演示联系人的会话', now()) RETURNING id`, tenantID, acctID).Scan(&convID); err != nil {
		t.Fatalf("conversation: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO messages(tenant_id, source_account_id, conversation_id, sender_contact_id, direction, message_type, content_text, sent_at)
		VALUES ($1,$2,$3,$4,'incoming','text','你好，这是验收消息', now())`,
		tenantID, acctID, convID, contactID); err != nil {
		t.Fatalf("message: %v", err)
	}
	return token, ts, convID
}

func TestEnterpriseContacts(t *testing.T) {
	token, ts, _ := seedEnterpriseData(t)
	code, body := doJSON(t, "GET", ts.URL+"/api/v1/contacts?q=演示", nil, token)
	if code != http.StatusOK {
		t.Fatalf("contacts status %d", code)
	}
	m := body.(map[string]any)
	if m["total"].(float64) < 1 {
		t.Fatalf("expected at least 1 contact, got %v", m["total"])
	}
}

func TestEnterpriseConversations(t *testing.T) {
	token, ts, _ := seedEnterpriseData(t)
	code, body := doJSON(t, "GET", ts.URL+"/api/v1/conversations", nil, token)
	if code != http.StatusOK {
		t.Fatalf("conversations status %d", code)
	}
	if body.(map[string]any)["total"].(float64) < 1 {
		t.Fatalf("expected conversations, got %v", body)
	}
}

func TestEnterpriseMessages(t *testing.T) {
	token, ts, convID := seedEnterpriseData(t)
	code, body := doJSON(t, "GET", ts.URL+"/api/v1/messages?conversation_id="+convID, nil, token)
	if code != http.StatusOK {
		t.Fatalf("messages status %d", code)
	}
	if body.(map[string]any)["total"].(float64) < 1 {
		t.Fatalf("expected messages, got %v", body)
	}
}

var _ = sql.ErrNoRows // 占位引用