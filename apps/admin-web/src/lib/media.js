// 媒体协议 schema 前端消费工具（R42.7）。
// 语义对齐 packages/protocol/media-manifest.json 与 media-receipt.json：
//  - manifest 必填字段、sha256 hex(64)、media_type/state 枚举、origin 必填
//  - receipt 必填字段、seq 十进制字符串、object_ref/media_kind 成对出现
//  - ACK 仅在耐久提交后发出；receipt 不代表已发布/索引/媒体已验证
//  - 对账：同一对象在已持久化(manifest 存在)前不能有收据；同对象不允许重复新增
// 纯函数、无副作用；只消费报告对象，不篡改数据。

const MEDIA_TYPES = ['image', 'video', 'audio', 'file'];
const MANIFEST_STATES = ['stored', 'deleting', 'deleted', 'quarantined'];
const HEX64 = /^[a-f0-9]{64}$/;

export function manifestErrors(m) {
  if (!m || typeof m !== 'object') return ['manifest 缺失'];
  const errs = [];
  const req = ['manifest_id', 'catalog_version', 'tenant_id', 'deployment_id', 'object_ref', 'media_type', 'origin', 'bytes_ref', 'sha256', 'size_bytes', 'seq', 'created_at', 'state'];
  for (const k of req) if (m[k] === undefined) errs.push(`缺少必填字段 ${k}`);
  if (m.sha256 !== undefined && !HEX64.test(m.sha256)) errs.push('sha256 必须是 64 位 hex');
  if (m.media_type !== undefined && !MEDIA_TYPES.includes(m.media_type)) errs.push('media_type 非法');
  if (m.state !== undefined && !MANIFEST_STATES.includes(m.state)) errs.push('state 非法');
  const o = m.origin;
  if (o !== undefined) {
    if (typeof o !== 'object') errs.push('origin 必须是对象');
    else for (const k of ['kind', 'source', 'source_message_ref']) if (o[k] === undefined) errs.push(`origin 缺少 ${k}`);
    if (o && o.kind !== undefined && !['original', 'derived'].includes(o.kind)) errs.push('origin.kind 非法');
  }
  return errs;
}

export function receiptErrors(r) {
  if (!r || typeof r !== 'object') return ['receipt 缺失'];
  const errs = [];
  const req = ['receipt_id', 'catalog_version', 'tenant_id', 'deployment_id', 'source', 'source_message_ref', 'seq', 'committed_at', 'backup_set'];
  for (const k of req) if (r[k] === undefined) errs.push(`缺少必填字段 ${k}`);
  if (r.seq !== undefined && !(typeof r.seq === 'string' && /^\d+$/.test(r.seq))) errs.push('seq 必须是十进制字符串');
  if (r.media_kind !== undefined && !MEDIA_TYPES.includes(r.media_kind)) errs.push('media_kind 非法');
  // object_ref 与 media_kind 必须成对出现（非媒体消息收据不带对象）
  if ((r.object_ref === undefined) !== (r.media_kind === undefined)) {
    errs.push('object_ref 与 media_kind 必须同时出现或同时缺失');
  }
  return errs;
}

// 对账：校验 receipt 与 manifest 的对应关系。
// 返回 { ok, errors }；ok=false 时不能对外发"已接受"确认（不发假 ACK）。
export function reconcile(receipt, manifest) {
  const rErr = receiptErrors(receipt);
  if (rErr.length) return { ok: false, errors: rErr };
  // 非媒体收据（无对象）无需 manifest 对账
  if (receipt.object_ref === undefined) return { ok: true, errors: [] };
  if (!manifest) return { ok: false, errors: ['收据引用了对象但无对应 manifest：拒绝假 ACK'] };
  const mErr = manifestErrors(manifest);
  if (mErr.length) return { ok: false, errors: mErr };
  if (manifest.object_ref !== receipt.object_ref) {
    return { ok: false, errors: ['manifest.object_ref 与收据不一致'] };
  }
  // 仅 stored 状态可确认接收；deleting/deleted/quarantined 都表示对象不可作为
  // 已持久化可接收状态（receipt 是已耐久接受的证明，对象已删/隔离则不 ACK）。
  if (manifest.state !== 'stored') {
    return { ok: false, errors: [`对象处于 ${manifest.state}，不能确认接收`] };
  }
  return { ok: true, errors: [] };
}