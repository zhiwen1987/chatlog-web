import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import test from 'node:test';
import assert from 'node:assert/strict';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const code = await readFile(path.join(root, 'src/lib/license.js'), 'utf8');
const m = await import('data:text/javascript;base64,' + Buffer.from(code).toString('base64'));

const NOW = Date.parse('2026-03-01T00:00:00Z');
const catalog = {
  'archive.read': { dependsOn: [] },
  'archive.ingest': { dependsOn: [] },
  'media.upload.image': { dependsOn: ['archive.ingest'] },
};
const grants = (over) => ([
  { grant_id: 'g1', feature_key: 'archive.read', state: 'active', valid_from: '2026-01-01T00:00:00Z', valid_until: '2027-01-01T00:00:00Z' },
  { grant_id: 'g2', feature_key: 'archive.ingest', state: 'active', valid_from: '2026-01-01T00:00:00Z', valid_until: '2027-01-01T00:00:00Z' },
  ...(over || []),
]);

test('grantActive accepts left-closed right-open active grant', () => {
  const g = grants()[0];
  assert.equal(m.grantActive(g, NOW), true);
  assert.equal(m.grantActive(g, Date.parse('2027-01-01T00:00:00Z')), false); // right-open boundary
  assert.equal(m.grantActive(g, Date.parse('2025-12-31T23:59:59Z')), false);
});

test('grantActive rejects non-active states and malformed windows', () => {
  assert.equal(m.grantActive({ ...grants()[0], state: 'paused' }, NOW), false);
  assert.equal(m.grantActive({ ...grants()[0], state: 'revoked' }, NOW), false);
  assert.equal(m.grantActive({ ...grants()[0], valid_from: 'not-a-date' }, NOW), false);
  assert.equal(m.grantActive(null, NOW), false);
});

test('dependencyCheck requires feature and all depends_on active', () => {
  const gs = grants([{ grant_id: 'g3', feature_key: 'media.upload.image', state: 'active', valid_from: '2026-01-01T00:00:00Z', valid_until: '2027-01-01T00:00:00Z' }]);
  const ok = m.dependencyCheck('media.upload.image', gs, catalog, NOW);
  assert.equal(ok.available, true);
  assert.deepEqual(ok.missing, []);
});

test('dependencyCheck reports missing when a dependency grant is inactive', () => {
  const gs = grants([{ grant_id: 'g3', feature_key: 'media.upload.image', state: 'active', valid_from: '2026-01-01T00:00:00Z', valid_until: '2027-01-01T00:00:00Z' }]);
  // 把依赖 archive.ingest 的 grant 置为 paused，模拟依赖不可用
  const gsPaused = gs.map((g) => g.feature_key === 'archive.ingest' ? { ...g, state: 'paused' } : g);
  const r = m.dependencyCheck('media.upload.image', gsPaused, catalog, NOW);
  assert.equal(r.available, false);
  assert.deepEqual(r.missing, ['archive.ingest']);
});

test('renderClaims marks v1/empty claims as not trusted (v2-only, no downgrade fake)', () => {
  const r = m.renderClaims({ features: ['archive.read'] }, catalog, NOW);
  assert.equal(r.trusted, false);
  assert.equal(r.grants.length, 0);
  assert.ok(r.note && r.note.includes('v2'));
  assert.equal(m.renderClaims(null, catalog, NOW).trusted, false);
});

test('renderClaims lists grants with active flag', () => {
  const gs = grants([{ grant_id: 'g3', feature_key: 'media.upload.image', state: 'active', valid_from: '2026-01-01T00:00:00Z', valid_until: '2027-01-01T00:00:00Z' }]);
  const r = m.renderClaims({ feature_grants: gs }, catalog, NOW);
  assert.equal(r.trusted, true);
  assert.equal(r.grants.length, 3);
  assert.equal(r.grants.find((g) => g.featureKey === 'archive.read').active, true);
  assert.equal(r.grants.find((g) => g.featureKey === 'media.upload.image').usable, true);
});
test('dependencyCheck does not infinitely recurse on dependency cycle', () => {
  const cyc = { 'a': { dependsOn: ['b'] }, 'b': { dependsOn: ['a'] } };
  const gs = [
    { grant_id: 'c1', feature_key: 'a', state: 'active', valid_from: '2026-01-01T00:00:00Z', valid_until: '2027-01-01T00:00:00Z' },
    { grant_id: 'c2', feature_key: 'b', state: 'active', valid_from: '2026-01-01T00:00:00Z', valid_until: '2027-01-01T00:00:00Z' },
  ];
  const r = m.dependencyCheck('a', gs, cyc, NOW);
  assert.equal(r.available, false);
  assert.ok(r.missing.includes('a'), 'cycle node should be reported as missing');
});
