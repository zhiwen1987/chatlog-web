package handler

// license.go — 许可状态只读端点（R42.8）。
// GET /api/v1/license/status：返回当前许可 claims v2 的授权状态摘要。
// 只做展示判定（CheckFeature）。配置了 LicenseVerifyKey 时实时查库并 JWS 验签，
// 验签失败/篡改按默认拒绝；未配置保持直接解析（兼容现状）。

import (
	"net/http"
	"time"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/db"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

// licenseStatus 返回许可状态摘要。每次实时查 DB（deployment 最新 revision）；
// 查不到或出错时默认拒绝（R42.7 默认拒绝）。
func (s *Server) licenseStatus(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	var c *model.Claims
	if s.DB != nil {
		var err error
		if len(s.LicenseVerifyKey) > 0 {
			c, err = db.LoadLicenseClaimsVerified(r.Context(), s.DB, s.DeploymentID, s.LicenseVerifyKey, s.LicenseExpectedAud)
		} else {
			c, err = db.LoadLicenseClaims(r.Context(), s.DB, s.DeploymentID)
		}
		if err != nil {
			// 验签失败/篡改按默认拒绝：不暴露签名细节，返回 present=false 语义。
			writeErr(w, http.StatusInternalServerError, "license lookup failed")
			return
		}
	} else {
		c = s.License // 单测/无 DB 环境回退启动快照
	}
	resp := map[string]any{
		"present": c != nil,
		"mode":    nil,
		"checked": now.Format(time.RFC3339),
	}
	if c != nil {
		resp["licensee"] = c.Licensee
		resp["deployment"] = c.Deployment
	}
	// 关键 feature 判定（与前端 ContractsView 消费的语义一致）
	keys := []string{"archive.read", "archive.ingest", "media.upload.image", "media.preview"}
	feat := map[string]auth.FeatureDecision{}
	for _, k := range keys {
		feat[k] = auth.CheckFeature(c, k, now)
	}
	resp["features"] = feat
	writeJSON(w, http.StatusOK, resp)
}
