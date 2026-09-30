package handler

// ingest.go — 数据链接入端点（R42 数据链 ingestion）。
// POST /api/v1/ingest/messages：租户内单条消息接入，幂等（upstream_message_id 去重）。
// 授权：要求 license claims 的 archive.ingest 可用（默认拒绝，R42.8）；
// 认证由 authed 中间件保证。ACK 语义：仅在该行耐久提交后返回 accepted>0。

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/db"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

var validSourceTypes = map[string]bool{
	"chatlog_http": true, "plaintext_wechat_sqlite": true, "manual_file_import": true,
	"windows_wechat": true, "macos_wechat": true, "linux_wechat": true, "enterprise_wechat": true,
}
var validDirections = map[string]bool{"incoming": true, "outgoing": true, "unknown": true}
var validMessageTypes = map[string]bool{"text": true, "image": true, "video": true, "voice": true, "file": true, "emoji": true, "system": true, "link": true, "quote": true, "unknown": true}
var validDecodeStatus = map[string]bool{"ok": true, "partial": true, "failed": true, "encrypted": true}

func (s *Server) ingestMessages(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ContextTenantID(r.Context())
	if tenantID == "" {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	// 授权门禁：archive.ingest 未授权默认拒绝（R42.8）。
	dec := auth.CheckFeature(s.License, model.FeatureArchiveIngest, time.Now().UTC())
	if !dec.Allowed {
		writeErr(w, http.StatusForbidden, "archive.ingest not licensed")
		return
	}
	var m model.IngestMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if m.UpstreamMessageID == "" || m.SourceType == "" || m.SourceExternalID == "" || m.ConversationRef == "" {
		writeErr(w, http.StatusBadRequest, "upstream_message_id/source_type/source_external_id/conversation_ref required")
		return
	}
	if !validSourceTypes[m.SourceType] {
		writeErr(w, http.StatusBadRequest, "source_type 非法")
		return
	}
	if !validDirections[m.Direction] || !validMessageTypes[m.MessageType] || !validDecodeStatus[m.DecodeStatus] {
		writeErr(w, http.StatusBadRequest, "direction/message_type/decode_status 枚举非法")
		return
	}
	accepted, duplicated, err := db.IngestMessage(r.Context(), s.DB, tenantID, &m)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "ingest failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"accepted": accepted, "duplicated": duplicated,
	})
}