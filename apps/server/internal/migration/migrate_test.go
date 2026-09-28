package migration

import (
	"context"
	"os"
	"testing"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/db"
)

// TestMigrateIntegration 需要 TEST_DATABASE_URL 指向可用 PostgreSQL（如 Docker）。
// 幂等验证：连续执行两次不报错。
func TestMigrateIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	conn, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	if err := Migrate(ctx, conn); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	if err := Migrate(ctx, conn); err != nil {
		t.Fatalf("second migrate (idempotency): %v", err)
	}

	// 校验核心表存在
	tables := []string{
		"tenants", "users", "tenant_members", "devices",
		"source_accounts", "source_shards", "contacts", "customers",
		"customer_identities", "conversations", "messages", "message_observations",
		"audit_log", "schema_migrations",
	}
	for _, tb := range tables {
		var n int
		if err := conn.QueryRowContext(ctx,
			`SELECT count(*) FROM information_schema.tables WHERE table_name = $1`, tb).Scan(&n); err != nil {
			t.Fatalf("query table %s: %v", tb, err)
		}
		if n == 0 {
			t.Fatalf("table %s missing after migration", tb)
		}
	}
}