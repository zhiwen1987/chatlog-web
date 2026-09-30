import { readFile } from 'node:fs/promises';
import { fileURLToPath, pathToFileURL } from 'node:url';
import path from 'node:path';
import test from 'node:test';
import assert from 'node:assert/strict';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const code = await readFile(path.join(root, 'src/lib/contracts.js'), 'utf8');
const m = await import(pathToFileURL(path.join(root, 'src/lib/contracts.js')).href);

test('aggregates all four tool namespaces', () => {
  for (const fn of ['countErrors', 'statusLabel', 'renderReport',
    'grantActive', 'dependencyCheck', 'renderClaims',
    'manifestErrors', 'receiptErrors', 'reconcile',
    'decimalStringError', 'checkDecimalFields', 'timeUnknownError', 'ownerReport']) {
    assert.equal(typeof m[fn], 'function', `missing export ${fn}`);
  }
});

test('validateContract dispatches integrity-report counts', () => {
  const ok = m.validateContract('integrity-report', { counts: { total_discovered: 10, in_scope: 8, verified: 8, pending: 0, excluded: 0, source_missing: 0 } });
  assert.equal(ok.ok, true);
  const bad = m.validateContract('integrity-report', { counts: { total_discovered: 10, in_scope: 8, verified: 9, pending: 0, excluded: 0, source_missing: 0 } });
  assert.equal(bad.ok, false);
});

test('validateContract dispatches media-manifest errors', () => {
  const bad = m.validateContract('media-manifest', { sha256: 'nope' });
  assert.equal(bad.ok, false);
  assert.ok(bad.errors.some((e) => e.includes('sha256')));
});

test('validateContract dispatches media-receipt', () => {
  const ok = m.validateContract('media-receipt', { receipt_id: 'r', catalog_version: 1, tenant_id: 't', deployment_id: 'd', source: 's', source_message_ref: 'm', seq: '1', committed_at: '2026-01-01T00:00:00Z', backup_set: 'b' });
  assert.equal(ok.ok, true);
});

test('validateContract rejects unknown type', () => {
  const r = m.validateContract('no-such-type', {});
  assert.equal(r.ok, false);
  assert.ok(r.errors[0].includes('未知契约类型'));
});

test('validateContract license-claims without v2 grants is not trusted', () => {
  const r = m.validateContract('license-claims', { features: ['archive.read'] });
  assert.equal(r.ok, false);
  assert.ok(r.errors[0].includes('v2 feature_grants'));
});
test('validateContract dispatches source-adapter', () => {
  const bad = m.validateContract('source-adapter', { source_id: 'x', cursor: { watermark: 42 } });
  assert.equal(bad.ok, false);
  assert.ok(bad.errors.some((e) => e.includes('watermark')));
});
