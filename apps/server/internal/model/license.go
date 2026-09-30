package model

// license.go — license-claims v2 解析与功能授权检查（R42.2/R42.8）。
// 语义对齐 packages/contracts/license-claims.json：
//  - feature_grants[] 是商业权利唯一来源；外层 licensee/deployment/公钥绑定等
//    由签发方保证，本文件只做结构解析与 grant 有效性
//  - grant state 枚举：active/paused/revoked/expired
//  - 依赖满足：feature 可用当且仅当自身 grant active 且全部 depends_on 可用
//  - 仅结构/语义检查，不验证 JWS 签名（需真实签发方私钥，R42.8）

import (
	"time"
)

// GrantState grant 状态枚举（对齐 license-claims.json state enum）。
type GrantState string

const (
	GrantActive  GrantState = "active"
	GrantPaused  GrantState = "paused"
	GrantRevoked GrantState = "revoked"
	GrantExpired GrantState = "expired"
)

// Grant 单个功能授权（对齐 feature_grants 元素）。
type Grant struct {
	GrantID    string     `json:"grant_id"`
	FeatureKey string     `json:"feature_key"`
	State      GrantState `json:"state"`
	ValidFrom  string     `json:"valid_from"`
	ValidUntil string     `json:"valid_until,omitempty"`
}

// Claims license-claims v2 外层结构（只解析授权相关字段）。
type Claims struct {
	Typ               string  `json:"typ"`
	Iss               string  `json:"iss"`
	Aud               string  `json:"aud"`
	Licensee          string  `json:"licensee"`
	Deployment        string  `json:"deployment"`
	FeatureCatalogVer int     `json:"feature_catalog_version"`
	LicenseRevision   int     `json:"license_revision"`
	FeatureGrants     []Grant `json:"feature_grants"`
}

// parseRFC3339OrZero 解析 ISO-8601；无效返回零值（配合 grantActive 判 false）。
func parseRFC3339OrZero(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// grantActive 判定 grant 在 now 时刻是否有效（active 且窗口左闭右开覆盖 now）。
func (g Grant) grantActive(now time.Time) bool {
	if g.State != GrantActive {
		return false
	}
	from := parseRFC3339OrZero(g.ValidFrom)
	if from.IsZero() {
		return false
	}
	if !now.Before(from) == false && now.Before(from) {
		return false
	}
	if now.Before(from) {
		return false
	}
	if g.ValidUntil != "" {
		until := parseRFC3339OrZero(g.ValidUntil)
		if until.IsZero() {
			return false
		}
		if !now.Before(until) { // 右开：now < until
			return false
		}
	}
	return true
}

// FeatureDeps 依赖表由 scripts/gen_feature_types.py 从 feature-catalog.json 生成
// （见 features.go，R42.2 单一来源）。此处复用，不重复定义。
// grantFor 返回 featureKey 的 grant；未找到返回 nil。
func (c *Claims) grantFor(featureKey string) *Grant {
	for i := range c.FeatureGrants {
		if c.FeatureGrants[i].FeatureKey == featureKey {
			return &c.FeatureGrants[i]
		}
	}
	return nil
}

// hasFeatureRec 递归检查 feature 及其依赖（visited 防环）。
func (c *Claims) hasFeatureRec(key string, visited map[string]bool, now time.Time) bool {
	if visited[key] {
		return false // 依赖环：视为不满足，防无限递归
	}
	visited[key] = true
	g := c.grantFor(key)
	if g == nil || !g.grantActive(now) {
		return false
	}
	for _, dep := range FeatureDeps[key] {
		if !c.hasFeatureRec(dep, visited, now) {
			return false
		}
	}
	return true
}

// HasFeature 检查 feature 是否已授权（含依赖满足）。now 可传 time.Now()。
func (c *Claims) HasFeature(featureKey string, now time.Time) bool {
	if c == nil || len(c.FeatureGrants) == 0 {
		return false
	}
	return c.hasFeatureRec(featureKey, map[string]bool{}, now)
}

// MissingDeps 返回 feature 未满足的直接依赖（供诊断展示；空=已满足）。
func (c *Claims) MissingDeps(featureKey string, now time.Time) []string {
	if c == nil {
		return nil
	}
	var missing []string
	for _, dep := range FeatureDeps[featureKey] {
		if !c.HasFeature(dep, now) {
			missing = append(missing, dep)
		}
	}
	return missing
}
