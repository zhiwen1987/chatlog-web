package handler

// license_write.go — 许可 claims 写入端点（R42.8，发行方/后台）。
// POST /api/v1/license：限 owner/admin；接收 license-claims v2 JSON，
// 校验可解析后写入 license_claims 表（幂等 upsert by deployment）。
// 只做结构校验，不验证 JWS 签名（R42.8：需真实签发方）。

import (
	"encoding/json"
	"net/http"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

type licenseWriteRequest struct {
	DeploymentID string          `json:"deployment_id"`
	Claims       json.RawMessage `json:"claims"`
}

func (s *Server) writeLicense(w http.ResponseWriter, r *http.Request) {
	if !hasRole(r, "owner", "admin") {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	var req licenseWriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.DeploymentID == "" || len(req.Claims) == 0 {
		writeErr(w, http.StatusBadRequest, "deployment_id and claims required")
		return
	}
	// 结构校验：claims 必须可解析为 model.Claims
	var c model.Claims
	if err := json.Unmarshal(req.Claims, &c); err != nil {
		writeErr(w, http.StatusBadRequest, "claims must be license-claims v2 JSON")
		return
	}
	if len(c.FeatureGrants) == 0 {
		writeErr(w, http.StatusBadRequest, "claims must contain feature_grants")
		return
	}

	// 幂等 upsert：同 deployment 更新为最新 revision（revision 由请求 claims 提供）
	revision := c.LicenseRevision
	if revision < 1 {
		revision = 1
	}
	_, err := s.DB.ExecContext(r.Context(), `
		INSERT INTO license_claims (licensee, deployment_id, claims_json, license_revision, updated_at)
		VALUES ($1, $2, $3::jsonb, $4, now())
		ON CONFLICT (licensee, deployment_id) DO UPDATE SET
			claims_json = EXCLUDED.claims_json,
			license_revision = EXCLUDED.license_revision,
			updated_at = now()`,
		c.Licensee, req.DeploymentID, string(req.Claims), revision)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "write failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
