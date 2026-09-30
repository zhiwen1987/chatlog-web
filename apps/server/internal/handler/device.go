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