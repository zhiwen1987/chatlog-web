package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// Entry 一条审计记录。
type Entry struct {
	TenantID   string
	UserID     string
	DeviceID   string
	Action     string
	Resource   string
	ResourceID string
	Detail     map[string]any
	IPAddress  string
	UserAgent  string
}

// Log 写入审计日志（失败不阻塞主流程，仅记录错误到标准输出）。
func Log(ctx context.Context, db *sql.DB, e Entry) {
	if db == nil {
		return
	}
	detail, err := json.Marshal(e.Detail)
	if err != nil {
		detail = []byte(`{}`)
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO audit_log
			(tenant_id, user_id, device_id, action, resource, resource_id, detail, ip_address, user_agent)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		nullStr(e.TenantID), nullStr(e.UserID), nullStr(e.DeviceID),
		e.Action, e.Resource, nullStr(e.ResourceID), detail, nullStr(e.IPAddress), nullStr(e.UserAgent),
	)
	if err != nil {
		// 审计失败不应影响业务；此处留给调用方观测。
		println("audit log error:", err.Error(), time.Now().Format(time.RFC3339))
	}
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}