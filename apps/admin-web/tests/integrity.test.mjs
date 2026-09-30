import {readFile} from 'node:fs/promises';
import {fileURLToPath} from 'node:url';
import path from 'node:path';
import test from 'node:test';
import assert from 'node:assert/strict';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const code = await readFile(path.join(root, 'src/lib/integrity.js'), 'utf8');
const m = await import('data:text/javascript;base64,' + Buffer.from(code).toString('base64'));

const validCounts = { total_discovered: 100, in_scope: 80, verified: 50, pending: 20, excluded: 8, source_missing: 2 };

test('valid counts produce no errors', () => {
  assert.deepEqual(m.countErrors(validCounts), []);
});

test('verified+pending+excluded+source_missing exceeding in_scope is rejected', () => {
  const errs = m.countErrors({ ...validCounts, verified: 60, pending: 30 });
  assert.ok(errs.some(e => e.includes('counts 不一致')));
});

test('in_scope exceeding total_discovered is rejected', () => {
  const errs = m.countErrors({ ...validCounts, in_scope: 120 });
  assert.ok(errs.some(e => e.includes('counts 不一致')));
});

test('missing or non-integer counts are rejected, not zero-filled', () => {
  assert.ok(m.countErrors(null).length >= 1);
  assert.ok(m.countErrors({ total_discovered: '100' }).some(e => e.includes('非负整数')));
});

test('excluded_by_license is labeled 未授权排除, source_missing is 源缺失', () => {
  assert.equal(m.statusLabel('excluded_by_license'), '未授权排除');
  assert.equal(m.statusLabel('source_missing'), '源缺失');
  assert.equal(m.statusLabel('paused_by_policy'), '策略暂停');
});

test('renderReport with zero in_scope does not fabricate completeness', () => {
  const r = m.renderReport({ counts: { ...validCounts, in_scope: 0, verified: 0, pending: 0, excluded: 0, source_missing: 0 } });
  assert.equal(r.completeness, null);
  assert.equal(r.trustable, true);
});

test('renderReport throws UNTRUSTED_COUNTS on inconsistent counts (no fake full report)', () => {
  assert.throws(
    () => m.renderReport({ counts: { ...validCounts, verified: 90, pending: 20 } }),
    (e) => e.code === 'UNTRUSTED_COUNTS'
  );
});

test('renderReport computes completeness and separates out-of-scope', () => {
  const r = m.renderReport({ counts: validCounts });
  assert.equal(r.completeness, 0.625);
  assert.equal(r.verified, 50);
  assert.equal(r.outOfScope, 20);
  assert.equal(r.total, 100);
});
test('renderReport rejects verified item with reason (fake completeness)', () => {
  assert.throws(
    () => m.renderReport({ counts: validCounts, items: [{ status: 'verified', reason: 'x' }] }),
    (e) => e.code === 'UNTRUSTED_COUNTS'
  );
  const ok = m.renderReport({ counts: validCounts, items: [{ status: 'verified' }] });
  assert.equal(ok.trustable, true);
});

test('negative counts are rejected (no silent negative completeness)', () => {
  assert.throws(
    () => m.renderReport({ counts: { total_discovered: -1, in_scope: 8, verified: 8, pending: 0, excluded: 0, source_missing: 0 } }),
    (e) => e.code === 'UNTRUSTED_COUNTS'
  );
});

test('non-integer counts are rejected, not truncated', () => {
  assert.ok(m.countErrors({ total_discovered: 10.5, in_scope: 8, verified: 8, pending: 0, excluded: 0, source_missing: 0 }).some((e) => e.includes('非负整数')));
});
