package handler

// media.go — 媒体清单/收据持久化与对账端点（R42.7 媒体协议）。
// POST /api/v1/media/manifests：上报 media-manifest（对齐 packages/protocol/media-manifest.json）。
// POST /api/v1/media/receipts：上报 media-receipt（对齐 packages/protocol/media-receipt.json）。
// GET  /api/v1/media/receipts：返回租户收据+对应清单列表（供前端 reconcile 对账，不发假 ACK）。
// 结构校验：必填字段非空 + sha256 64 位 hex + media_type/state 枚举（与前端 media.js 同语义）。

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/db"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

var (
	sha256HexRe  = regexp.MustCompile(`^[a-f0-9]{64}$`)
	mediaTypes   = map[string]bool{"image": true, "video": true, "audio": true, "file": true}
	manifestStts = map[string]bool{"stored": true, "deleting": true, "deleted": true, "quarantined": true}
)

func (s *Server) saveMediaManifest(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ContextTenantID(r.Context())
	if tenantID == "" {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	var m model.MediaManifest
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if m.ManifestID == "" || m.ObjectRef == "" || m.SHA256 == "" || m.MediaType == "" ||
		m.BytesRef == "" || m.Seq == "" || m.CreatedAt == "" || m.State == "" {
		writeErr(w, http.StatusBadRequest, "manifest 必填字段缺失")
		return
	}
	if !sha256HexRe.MatchString(m.SHA256) {
		writeErr(w, http.StatusBadRequest, "sha256 必须是 64 位 hex")
		return
	}
	if !mediaTypes[m.MediaType] {
		writeErr(w, http.StatusBadRequest, "media_type 非法")
		return
	}
	if !manifestStts[m.State] {
		writeErr(w, http.StatusBadRequest, "state 非法")
		return
	}
	if m.Origin.Source == "" || m.Origin.SourceMessageRef == "" {
		writeErr(w, http.StatusBadRequest, "origin 必填 source/source_message_ref")
		return
	}
	if err := db.SaveMediaManifest(r.Context(), s.DB, tenantID, &m); err != nil {
		writeErr(w, http.StatusInternalServerError, "save manifest failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "manifest_id": m.ManifestID})
}

func (s *Server) saveMediaReceipt(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ContextTenantID(r.Context())
	if tenantID == "" {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	var rc model.MediaReceipt
	if err := json.NewDecoder(r.Body).Decode(&rc); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if rc.ReceiptID == "" || rc.Source == "" || rc.SourceMessageRef == "" ||
		rc.Seq == "" || rc.CommittedAt == "" || rc.BackupSet == "" {
		writeErr(w, http.StatusBadRequest, "receipt 必填字段缺失")
		return
	}
	// object_ref 与 media_kind 必须成对出现
	if (rc.ObjectRef == "") != (rc.MediaKind == "") {
		writeErr(w, http.StatusBadRequest, "object_ref 与 media_kind 必须同时出现或同时缺失")
		return
	}
	if rc.MediaKind != "" && !mediaTypes[rc.MediaKind] {
		writeErr(w, http.StatusBadRequest, "media_kind 非法")
		return
	}
	if err := db.SaveMediaReceipt(r.Context(), s.DB, tenantID, &rc); err != nil {
		writeErr(w, http.StatusInternalServerError, "save receipt failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "receipt_id": rc.ReceiptID})
}

func (s *Server) listMediaReceipts(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ContextTenantID(r.Context())
	if tenantID == "" {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	rows, err := db.ListMediaReceiptsForReconcile(r.Context(), s.DB, tenantID, 50)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": rows,
		"total": len(rows),
	})
}