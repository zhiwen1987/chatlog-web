package handler

// device.go — 设备注册端点（R42.7 设备登记，presence/heartbeat 的前置依赖）。
// POST /api/v1/devices/register：在 authed 中间件下，用 JWT 上下文
// （tenant_id/user_id）注册设备；按 (tenant_id, device_fingerprint) 幂等
// upsert，返回稳定 device_id 供后续 presence/heartbeat 使用。

import (
	"encoding/json"
	"net/http"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/db"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

type deviceRegisterRequest struct {
	DeviceName        string `json:"device_name"`
	DeviceFingerprint string `json:"device_fingerprint"`
	Platform          string `json:"platform"`
	PlatformVersion   string `json:"platform_version"`
	Architecture      string `json:"architecture"`
	ClientVersion     string `json:"client_version"`
	PublicKey         string `json:"public_key"`
}

var validPlatforms = map[string]bool{"windows": true, "macos": true, "linux": true}
var validArchitectures = map[string]bool{"x86_64": true, "arm64": true}

func (s *Server) registerDevice(w http.ResponseWriter, r *http.Request) {
	var req deviceRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.DeviceFingerprint == "" || req.DeviceName == "" {
		writeErr(w, http.StatusBadRequest, "device_name and device_fingerprint required")
		return
	}
	if !validPlatforms[req.Platform] {
		writeErr(w, http.StatusBadRequest, "platform must be windows/macos/linux")
		return
	}
	if !validArchitectures[req.Architecture] {
		writeErr(w, http.StatusBadRequest, "architecture must be x86_64/arm64")
		return
	}

	tenantID := auth.ContextTenantID(r.Context())
	userID := auth.ContextUserID(r.Context())
	if tenantID == "" || userID == "" {
		writeErr(w, http.StatusUnauthorized, "auth context missing")
		return
	}

	d := &model.Device{
		TenantID:          tenantID,
		UserID:            userID,
		DeviceName:        req.DeviceName,
		DeviceFingerprint: req.DeviceFingerprint,
		Platform:          req.Platform,
		PlatformVersion:   req.PlatformVersion,
		Architecture:      req.Architecture,
		ClientVersion:     req.ClientVersion,
		PublicKey:         req.PublicKey,
		Status:            "active",
	}
	id, err := db.UpsertDevice(r.Context(), s.DB, d)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "device register failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "device_id": id, "status": "active"})
}

// heartbeat 处理设备心跳上报（W29 presence/heartbeat）。
// POST /api/v1/devices/{id}/heartbeat：复用 devices.last_seen_at 记录在线时间，
// 返回当前 last_seen_at；设备不存在或不属于本租户返回 404。
// 归属校验用 URL 的 device_id + JWT 的 tenant_id（用户 JWT 不携带设备 did），
// 不推进消息 ACK（心跳只表达设备在线，与数据收据分离，AGENTS A07）。
func (s *Server) heartbeat(w http.ResponseWriter, r *http.Request, deviceID string) {
	tenantID := auth.ContextTenantID(r.Context())
	if deviceID == "" || tenantID == "" {
		writeErr(w, http.StatusUnauthorized, "auth context missing")
		return
	}

	ok, err := db.TouchDevice(r.Context(), s.DB, tenantID, deviceID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "heartbeat failed")
		return
	}
	if !ok {
		writeErr(w, http.StatusNotFound, "device not found")
		return
	}

	// 返回更新后的 last_seen_at（同一事务性读取，供客户端校准）。
	var lastSeenAt string
	row := s.DB.QueryRowContext(r.Context(),
		`SELECT to_char(last_seen_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')
		 FROM devices WHERE id = $1 AND tenant_id = $2`, deviceID, tenantID)
	if err := row.Scan(&lastSeenAt); err != nil {
		writeErr(w, http.StatusInternalServerError, "heartbeat read failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "device_id": deviceID, "last_seen_at": lastSeenAt})
}