// SourceAdapter 契约前端消费工具（W21）。
// 语义对齐 packages/source-contract/source-adapter.json（R05/R42.7）：
//  - capabilities 检查：detect/schema_probe/stable_id/shard_generation/cursor
//  - cursor 校验：watermark 十进制字符串（R42.7），is_exhaustive 标志
//  - overwrite_policy 语义：低位回填/重复观察/账号切换
//  - 来源只读、毒数据隔离、不冒称可靠 CDC（docs 二次开发指南）
// 纯函数、无副作用。

const CAP_KEYS = ['detect', 'schema_probe', 'stable_id', 'shard_generation', 'cursor'];
const DECIMAL_STR = /^\d+$/;

// 校验 source-adapter 结构：返回错误数组（空=通过）。
export function sourceErrors(s) {
  if (!s || typeof s !== 'object') return ['source 缺失'];
  const errs = [];
  const req = ['contract_version', 'source_id', 'capabilities', 'identity', 'cursor', 'unsupported', 'overwrite_policy'];
  for (const k of req) if (s[k] === undefined) errs.push(`缺少必填字段 ${k}`);
  if (s.capabilities !== undefined) {
    for (const k of CAP_KEYS) if (typeof s.capabilities[k] !== 'boolean') errs.push(`capabilities.${k} 必须是 boolean`);
  }
  const c = s.cursor;
  if (c !== undefined) {
    if (!(typeof c.watermark === 'string' && DECIMAL_STR.test(c.watermark))) errs.push('cursor.watermark 必须是十进制字符串');
    if (typeof c.is_exhaustive !== 'boolean') errs.push('cursor.is_exhaustive 必须是 boolean');
  }
  const ow = s.overwrite_policy;
  if (ow !== undefined) {
    for (const k of ['low_water_backfill', 'duplicate_observation', 'account_switch']) {
      if (typeof ow[k] !== 'string' || ow[k] === '') errs.push(`overwrite_policy.${k} 必须是非空字符串`);
    }
  }
  return errs;
}

// 能力检查：返回 { supported: boolean, missing: string[] }。
// missing 列出该来源缺失的能力（如无 cursor 则须全量对账）。
export function capabilitiesReport(s) {
  if (!s || typeof s.capabilities !== 'object') return { supported: false, missing: CAP_KEYS.slice() };
  const missing = CAP_KEYS.filter((k) => s.capabilities[k] !== true);
  return { supported: missing.length === 0, missing };
}

// 游标/覆盖策略展示：非 exhaustive 时标记须全量对账（不冒称可靠 CDC）。
export function syncPlan(s) {
  if (!s) return null;
  const cursor = s.cursor || {};
  const plan = {
    exhaustive: cursor.is_exhaustive === true,
    needFullReconcile: cursor.is_exhaustive !== true,
    backfill: (s.overwrite_policy && s.overwrite_policy.low_water_backfill) || null,
    duplicate: (s.overwrite_policy && s.overwrite_policy.duplicate_observation) || null,
    accountSwitch: (s.overwrite_policy && s.overwrite_policy.account_switch) || null,
  };
  return plan;
}