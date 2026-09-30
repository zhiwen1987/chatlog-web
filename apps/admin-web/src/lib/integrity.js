// 完整性报告前端消费工具（R42.10）。
// 语义对齐 packages/contracts/integrity-report.json：
//  - counts 一致性：verified+pending+excluded+source_missing <= in_scope <= total_discovered
//  - 三态：excluded_by_license（未授权）/ paused_by_policy（企业停用）/ source_missing（源缺失）
//  - 未授权不能当 missing，verified 更不能带 reason 冒充完整
// 纯函数、无副作用；只消费报告对象，不篡改数据。

const COUNT_KEYS = [
  'total_discovered', 'in_scope', 'verified', 'pending', 'excluded', 'source_missing'
];

const STATUS_LABELS = {
  verified: '已校验',
  pending: '待传输',
  excluded_by_license: '未授权排除',
  paused_by_policy: '策略暂停',
  source_missing: '源缺失',
};

export function countErrors(counts) {
  if (!counts || typeof counts !== 'object') return ['counts 缺失'];
  const errs = [];
  for (const k of COUNT_KEYS) {
    if (!Number.isInteger(counts[k])) errs.push(`${k} 必须是非负整数`);
  }
  if (errs.length) return errs;
  const { total_discovered: t, in_scope: s, verified: v, pending: p, excluded: e, source_missing: m } = counts;
  if (v + p + e + m > s) errs.push(`counts 不一致：verified+pending+excluded+source_missing(${v}+${p}+${e}+${m}) 超过 in_scope(${s})`);
  if (s > t) errs.push(`counts 不一致：in_scope(${s}) 超过 total_discovered(${t})`);
  return errs;
}

export function statusLabel(status) {
  return STATUS_LABELS[status] || status;
}

// 返回完整性报告的前端展示结构；counts 不满足一致性时抛出（不显示假完整）。
export function renderReport(report) {
  const counts = report && report.counts;
  const errs = countErrors(counts);
  // verified 项带 reason 是伪造完整的信号（R42.10：verified 不可能有排除原因）。
  for (const it of (report && report.items) || []) {
    if (it.status === 'verified' && it.reason) {
      errs.push('verified 项不能带 reason（不冒充完整）');
      break;
    }
  }
  if (errs.length) {
    const e = new Error(`integrity-report 不可信：${errs.join('；')}`);
    e.code = 'UNTRUSTED_COUNTS';
    throw e;
  }
  const { total_discovered: t, in_scope: s, verified: v, pending: p, excluded: e, source_missing: m } = counts;
  return {
    trustable: true,
    completeness: s === 0 ? null : v / s,          // 0..1；s=0 时不编造 0
    verified: v,
    pending: p,
    excluded: e,
    sourceMissing: m,
    outOfScope: Math.max(0, t - s),                 // 范围外（total 中不在应传范围的部分）
    total: t,
  };
}