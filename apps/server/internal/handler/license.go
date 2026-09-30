package handler

// license.go — 许可状态只读端点（R42.8）。
// GET /api/v1/license/status：返回当前许可 claims v2 的授权状态摘要。
// 只做展示判定（CheckFeature），不验证 JWS 签名（R42.8：需签发方）。

import (
	"net/http"
	"time"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
)

// licenseStatus 返回许可状态摘要。License 为 nil 时默认拒绝（R42.7 默认拒绝）。
func (s *Server) licenseStatus(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	c := s.License
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