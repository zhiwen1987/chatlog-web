package auth

// feature.go — 功能授权判定（R42.2/R42.8）。
// 纯函数：给定 license-claims v2（model.Claims）+ featureKey，判定是否授权。
// 依赖 model.HasFeature（feature_grants 唯一来源 + 依赖满足）。
// 不绑定数据源：Claims 由上层（DB/签发方）提供，本层只做判定。

import (
	"time"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

// FeatureDecision 功能授权判定结果。
type FeatureDecision struct {
	Allowed bool     // 是否允许
	Missing []string // 未满足的直接依赖（Allowed=false 时给出诊断）
}

// CheckFeature 判定 Claims 是否授予 featureKey。
//   - Claims 为 nil 或空 → 拒绝（默认拒绝，R42.7）
//   - 仅 structure/语义检查，不验证 JWS 签名（R42.8：需签发方）
func CheckFeature(c *model.Claims, featureKey string, now time.Time) FeatureDecision {
	if c == nil {
		return FeatureDecision{Allowed: false, Missing: nil}
	}
	if !c.HasFeature(featureKey, now) {
		return FeatureDecision{Allowed: false, Missing: c.MissingDeps(featureKey, now)}
	}
	return FeatureDecision{Allowed: true}
}
