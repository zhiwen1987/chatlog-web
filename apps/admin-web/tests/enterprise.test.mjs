// enterprise API 适配层行为测试（聚焦 getLicenseStatus）。
// 源文件用 vite alias（@/lib/data），node 直接加载会失败；
// 这里读真实源码，把 alias 替换为内联 data: URL 桩（getLicenseStatus 路径不调用 data 模块），
// 再以 data: URL import 真实方法逻辑。
import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import test from 'node:test';
import assert from 'node:assert/strict';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const code = await readFile(path.join(root, 'src/api/enterprise.js'), 'utf8');

// data 模块桩：仅保 import 可解析，getLicenseStatus 不调用其中任何函数
const stub = 'export const parseChatLogs=(x)=>x;export const normalizeContact=(x)=>x;' +
  'export const normalizeRoom=(x)=>x;export const normalizeSession=(x)=>x;' +
  'export const safeUrl=(u)=>/^https?:\\/\\//.test(u);export const csvString=(x)=>x';
const dataUrl = 'data:text/javascript,' + encodeURIComponent(stub);
const patched = code.replace("from '@/lib/data'", 'from ' + JSON.stringify(dataUrl));
const m = await import('data:text/javascript;base64,' + Buffer.from(patched).toString('base64'));

// 隔离每个测试的 localStorage
function freshLocalStorage () {
  const store = new Map();
  globalThis.localStorage = {
    getItem: (k) => (store.has(k) ? store.get(k) : null),
    setItem: (k, v) => { store.set(k, String(v)); },
    removeItem: (k) => { store.delete(k); },
  };
}

function jsonResponse (status, obj, headers = {}) {
  return {
    ok: status >= 200 && status < 300,
    status,
    headers: {
      get: (name) => (name.toLowerCase() === 'content-type' ? (headers['content-type'] || 'application/json') : null),
      entries: () => Object.entries(headers),
    },
    async text () { return JSON.stringify(obj); },
  };
}

test('getLicenseStatus unwraps object response with present/features', async () => {
  freshLocalStorage();
  m.default.setEnterpriseToken('tok');
  const payload = {
    present: true, mode: 'v2', checked: '2026-09-30T12:00:00Z',
    licensee: 'acme', deployment: 'd-1',
    features: { 'archive.read': { allowed: true, missing: null } },
  };
  globalThis.fetch = async () => jsonResponse(200, payload);
  const r = await m.default.getLicenseStatus();
  assert.equal(r.present, true);
  assert.equal(r.licensee, 'acme');
  assert.equal(r.features['archive.read'].allowed, true);
});

test('getLicenseStatus rejects 401/403 as auth error', async () => {
  freshLocalStorage();
  m.default.setEnterpriseToken('bad');
  globalThis.fetch = async () => jsonResponse(401, { error: 'unauthorized' });
  await assert.rejects(() => m.default.getLicenseStatus(), /拒绝访问/);
});

test('getLicenseStatus rejects HTML body as not-data', async () => {
  freshLocalStorage();
  m.default.setEnterpriseToken('tok');
  globalThis.fetch = async () => ({
    ok: true, status: 200,
    headers: { get: () => 'text/html', entries: () => [] },
    async text () { return '<!doctype html><html></html>'; },
  });
  await assert.rejects(() => m.default.getLicenseStatus(), /网页/);
});

test('getLicenseStatus propagates non-ok status as HTTP error', async () => {
  freshLocalStorage();
  m.default.setEnterpriseToken('tok');
  globalThis.fetch = async () => jsonResponse(500, {});
  await assert.rejects(() => m.default.getLicenseStatus(), /HTTP 500/);
});
test('getIntegrityReport unwraps object with counts', async () => {
  freshLocalStorage();
  m.default.setEnterpriseToken('tok');
  globalThis.fetch = async () => jsonResponse(200, {
    counts: { total_discovered: 3, in_scope: 2, verified: 1, pending: 1, excluded: 1, source_missing: 0 },
    generated_at: '2026-09-30T12:00:00Z',
  });
  const r = await m.default.getIntegrityReport();
  assert.equal(r.counts.verified, 1);
  assert.equal(r.counts.pending, 1);
  assert.equal(r.counts.excluded, 1);
  assert.equal(r.generated_at, '2026-09-30T12:00:00Z');
});

test('getIntegrityReport rejects 401/403 as auth error', async () => {
  freshLocalStorage();
  m.default.setEnterpriseToken('bad');
  globalThis.fetch = async () => jsonResponse(401, { error: 'unauthorized' });
  await assert.rejects(() => m.default.getIntegrityReport(), /拒绝访问/);
});

test('getIntegrityReport rejects HTML body as not-data', async () => {
  freshLocalStorage();
  m.default.setEnterpriseToken('tok');
  globalThis.fetch = async () => ({
    ok: true, status: 200,
    headers: { get: () => 'text/html', entries: () => [] },
    async text () { return '<!doctype html><html></html>'; },
  });
  await assert.rejects(() => m.default.getIntegrityReport(), /网页/);
});

test('getIntegrityReport propagates non-ok status as HTTP error', async () => {
  freshLocalStorage();
  m.default.setEnterpriseToken('tok');
  globalThis.fetch = async () => jsonResponse(500, {});
  await assert.rejects(() => m.default.getIntegrityReport(), /HTTP 500/);
});
