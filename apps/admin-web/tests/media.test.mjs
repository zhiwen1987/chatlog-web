import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import test from 'node:test';
import assert from 'node:assert/strict';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const code = await readFile(path.join(root, 'src/lib/media.js'), 'utf8');
const m = await import('data:text/javascript;base64,' + Buffer.from(code).toString('base64'));

const manifest = {
  manifest_id: 'mm-1', catalog_version: 1, tenant_id: 't1', deployment_id: 'd1',
  object_ref: 'obj.1', media_type: 'image',
  origin: { kind: 'original', source: 'wechat-desktop', source_message_ref: 'src.1' },
  bytes_ref: 's3://b/obj.1', sha256: 'a'.repeat(64), size_bytes: 10, seq: '1', created_at: '2026-01-01T00:00:00Z', state: 'stored',
};
const receipt = {
  receipt_id: 'r1', catalog_version: 1, tenant_id: 't1', deployment_id: 'd1',
  source: 'wechat-desktop', source_message_ref: 'src.1', seq: '1', committed_at: '2026-01-01T00:00:00Z', backup_set: 'bs1',
  object_ref: 'obj.1', media_kind: 'image',
};

test('valid manifest has no errors', () => assert.deepEqual(m.manifestErrors(manifest), []));

test('manifest rejects bad sha256, bad media_type, bad state, missing origin fields', () => {
  const errs = m.manifestErrors({ ...manifest, sha256: 'nope', media_type: 'gif', state: 'bogus', origin: { kind: 'original' } });
  assert.ok(errs.some((e) => e.includes('sha256')));
  assert.ok(errs.some((e) => e.includes('media_type')));
  assert.ok(errs.some((e) => e.includes('state')));
  assert.ok(errs.some((e) => e.includes('origin 缺少')));
});

test('valid receipt has no errors', () => assert.deepEqual(m.receiptErrors(receipt), []));

test('receipt rejects non-decimal seq and unpaired object_ref/media_kind', () => {
  assert.ok(m.receiptErrors({ ...receipt, seq: 42 }).some((e) => e.includes('十进制')));
  assert.ok(m.receiptErrors({ ...receipt, object_ref: undefined }).some((e) => e.includes('同时出现')));
  assert.ok(m.receiptErrors({ ...receipt, media_kind: undefined }).some((e) => e.includes('同时出现')));
});

test('reconcile ok when receipt and manifest agree', () => {
  const r = m.reconcile(receipt, manifest);
  assert.equal(r.ok, true);
});

test('reconcile rejects fake ACK: receipt references object without manifest', () => {
  const r = m.reconcile(receipt, null);
  assert.equal(r.ok, false);
  assert.ok(r.errors.some((e) => e.includes('拒绝假 ACK')));
});

test('reconcile rejects quarantined object', () => {
  const r = m.reconcile(receipt, { ...manifest, state: 'quarantined' });
  assert.equal(r.ok, false);
  assert.ok(r.errors.some((e) => e.includes('quarantined')));
});

test('non-media receipt (no object) reconciles without manifest', () => {
  const r = m.reconcile({ ...receipt, object_ref: undefined, media_kind: undefined }, null);
  assert.equal(r.ok, true);
});
test('reconcile rejects deleting/deleted states (only stored confirms receipt)', () => {
  for (const st of ['deleting', 'deleted']) {
    const r = m.reconcile(receipt, { ...manifest, state: st });
    assert.equal(r.ok, false, `state=${st} must not be confirmable`);
    assert.ok(r.errors.some((e) => e.includes('不能确认接收')));
  }
  const ok = m.reconcile(receipt, { ...manifest, state: 'stored' });
  assert.equal(ok.ok, true);
});
