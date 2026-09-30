// 商业许可 claims v2 前端消费工具（R42.8）。
// 语义对齐 packages/contracts/license-claims.json 与 feature-catalog.json：
//  - feature_grants[] 是商业权利唯一来源；旧 features[] 不在此消费
//  - grant state 枚举：active/paused/revoked/expired
//  - grant 窗口左闭右开；依赖（depends_on）须满足才视为可用
//  - v2-only 防降级：收到 v1 claims 不冒充 v2 激活
// 纯函数、无副作用。

const GRANT_STATES = ['active', 'paused', 'revoked', 'expired'];

function iso(s) {
  return typeof s === 'string' ? Date.parse(s) : Number.NaN;
}

// 返回该 grant 在 now 时刻是否有效（active 且窗口左闭右开覆盖 now）。
export function grantActive(grant, now = Date.now()) {
  if (!grant || !GRANT_STATES.includes(grant.state)) return false;
  if (grant.state !== 'active') return false;
  const from = iso(grant.valid_from);
  const to = grant.valid_until === undefined || grant.valid_until === null ? Infinity : iso(grant.valid_until);
  if (!Number.isFinite(from) || (grant.valid_until != null && !Number.isFinite(to))) return false;
  return from <= now && now < to;
}

// 依赖满足度：feature 可用当且仅当其自身 grant active 且全部 depends_on 可用。
// 返回 { available, missing }；missing 列出未满足的直接依赖 feature_key。
export function dependencyCheck(featureKey, grants, catalog, now = Date.now()) {
  const granted = (k) => grants.some((g) => g.feature_key === k && grantActive(g, now));
  const missing = [];
  const visiting = new Set();
  const visit = (k) => {
    if (visiting.has(k)) {
      // 依赖环（A→B→A）：视为依赖不满足，避免无限递归；feature-catalog 已保证
      // DAG 无环，但运行时可能拿到损坏目录，这里防御性防环。
      if (!missing.includes(k)) missing.push(k);
      return false;
    }
    visiting.add(k);
    let ok = granted(k);
    if (ok) {
      for (const dep of (catalog[k] && catalog[k].dependsOn) || []) {
        if (!visit(dep)) { ok = false; break; }
      }
    } else if (!missing.includes(k)) {
      missing.push(k);
    }
    visiting.delete(k);
    return ok;
  };
  const ok = visit(featureKey);
  return { available: ok, missing };
}

// 平台展示：grant 状态 + 依赖满足 + 防降级标记。counts 不涉及。
export function renderClaims(claims, catalog, now = Date.now()) {
  const out = { trusted: false, mode: null, grants: [] };
  if (!claims) return out;
  if (claims.feature_grants && !Array.isArray(claims.feature_grants)) return out;
  if (!claims.feature_grants) {
    // v1 或空 claims：不冒充 v2（R42.8 v2_only 防降级）
    out.note = '无 v2 feature_grants，不可作为激活凭证';
    return out;
  }
  out.mode = claims.mode || null;
  out.trusted = true;
  out.grants = claims.feature_grants.map((g) => {
    const active = grantActive(g, now);
    const dep = dependencyCheck(g.feature_key, claims.feature_grants, catalog, now);
    return {
      grantId: g.grant_id,
      featureKey: g.feature_key,
      state: g.state,
      active,
      depsSatisfied: dep.available,
      missingDeps: dep.missing,
      usable: active && dep.available,
    };
  });
  return out;
}