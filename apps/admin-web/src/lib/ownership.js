// 数据所有权与 ID 表示前端工具（R07/R39/R42.7）。
// 语义对齐 packages/contracts/data-ownership.json 的 field_mode：
//  - 64位ID/bytes/monotonic_seq/revision 跨 JS 必须用精确十进制字符串，不用浮点
//  - 时间字段分别命名；未知保留 null/状态，不填 0/当前时间/1970
//  - 不能把人类用户名、来源external ID、licensee ID、tenant ID、设备ID、
//    deployment ID 混成同一主键（id_kinds）
// 纯函数、无副作用；只消费数据对象，不篡改。

const DECIMAL_STR = /^\d+$/;
const FIELD_KINDS = ['int64_id', 'bytes', 'monotonic_seq', 'revision'];

// 十进制字符串表示校验：接受精确十进制字符串，拒绝浮点/负数/科学计数/非整数。
// 返回 '' 或错误信息。
export function decimalStringError(value, label = '值') {
  if (value === undefined || value === null || value === '') return '';
  if (typeof value === 'number') return `${label} 必须用十进制字符串（不能是浮点，防精度丢失）`;
  if (typeof value !== 'string') return `${label} 必须是字符串`;
  if (!DECIMAL_STR.test(value)) return `${label} 必须是纯十进制数字字符串`;
  return '';
}

// 校验一组字段的十进制字符串表示；fields: [{key, label}] 或 [key, ...]。
// 返回错误数组（空=通过）。
export function checkDecimalFields(obj, fields) {
  const errs = [];
  for (const f of fields) {
    const key = typeof f === 'string' ? f : f.key;
    const label = typeof f === 'string' ? f : (f.label || f.key);
    if (obj[key] === undefined || obj[key] === null || obj[key] === '') continue;
    const e = decimalStringError(obj[key], label);
    if (e) errs.push(e);
  }
  return errs;
}

// 未知时间不得变成今天/0/1970：接受 null/显式 unknown，拒绝字符串''被篡改。
export function timeUnknownError(value) {
  if (value === null || value === undefined || value === 'unknown') return '';
  if (value === '') return '未知时间不能伪装成空字符串';
  return '';
}

// 检查 owner 域清单：返回 { owners, unknownOwners, total }。
// unknownOwners 列出 data-ownership.json 未登记的 owner（防数据归属逃逸）。
export function ownerReport(rows, knownOwners) {
  const list = rows || [];
  const known = new Set(knownOwners || []);
  const unknown = [];
  for (const r of list) {
    const o = r && r.owner;
    if (o !== undefined && !known.has(o) && !unknown.includes(o)) unknown.push(o);
  }
  return { owners: list, unknownOwners: unknown, total: list.length };
}