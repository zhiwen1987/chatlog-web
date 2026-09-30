import { readFile } from 'node:fs/promises';
import { fileURLToPath, pathToFileURL } from 'node:url';
import path from 'node:path';
import test from 'node:test';
import assert from 'node:assert/strict';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const code = await readFile(path.join(root, 'src/lib/source.js'), 'utf8');
const m = await import('data:text/javascript;base64,' + Buffer.from(code).toString('base64'));

const valid = {
  contract_version: 1, source_id: 'wechat-desktop',
  capabilities: { detect: true, schema_probe: true, stable_id: true, shard_generation: true, cursor: true },
  identity: { kind: 'wxid', external_id: 'wxid_abc' },
  cursor: { watermark: '1000', is_exhaustive: true },
  unsupported: [],
  overwrite_policy: { low_water_backfill: '按消息身份去重', duplicate_observation: '标记重复', account_switch: '重置游标' },
};

test('valid source contract has no errors', () => assert.deepEqual(m.sourceErrors(valid), []));

test('source rejects bad cursor and capabilities', () => {
  const errs = m.sourceErrors({ ...valid, cursor: { watermark: 42, is_exhaustive: true }, capabilities: { detect: 'yes' } });
  assert.ok(errs.some((e) => e.includes('watermark')));
  assert.ok(errs.some((e) => e.includes('capabilities.detect')));
});

test('capabilitiesReport lists missing capabilities', () => {
  const r = m.capabilitiesReport({ capabilities: { detect: true, cursor: false } });
  assert.equal(r.supported, false);
  assert.ok(r.missing.includes('schema_probe'));
  assert.ok(r.missing.includes('cursor'));
});

test('syncPlan flags non-exhaustive cursor for full reconcile (no fake CDC)', () => {
  const r = m.syncPlan({ ...valid, cursor: { watermark: '1', is_exhaustive: false } });
  assert.equal(r.needFullReconcile, true);
  assert.equal(r.exhaustive, false);
  const full = m.syncPlan(valid);
  assert.equal(full.needFullReconcile, false);
  assert.equal(full.exhaustive, true);
});