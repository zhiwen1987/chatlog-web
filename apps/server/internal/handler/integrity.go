package handler

// integrity.go — 消息完整性报告只读端点（R42.10）。
// GET /api/v1/integrity/report：基于真实 messages 表返回租户消息完整性计数。
// 计数语义对齐前端 integrity.js 的 renderReport 消费（counts 一致性约束）。
// 租户隔离由 auth 中间件注入的 tenant_id + WHERE 过滤保证（与 contacts/messages 一致）。

import (
	"net/http"
	"time"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/db"
)

func (s *Server) integrityReport(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ContextTenantID(r.Context())
	if tenantID == "" {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	counts, err := db.LoadIntegrityReport(r.Context(), s.DB, tenantID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "integrity lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"counts": counts,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	})
}