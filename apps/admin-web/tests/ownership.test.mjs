import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import test from 'node:test';
import assert from 'node:assert/strict';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const code = await readFile(path.join(root, 'src/lib/ownership.js'), 'utf8');
const m = await import('data:text/javascript;base64,' + Buffer.from(code).toString('base64'));

test('decimal string accepts exact decimal string', () => {
  assert.equal(m.decimalStringError('9007199254740993', 'id'), '');
  assert.equal(m.decimalStringError('0001', 'seq'), '');
});

test('decimal string rejects float, negative, exponent and non-string', () => {
  assert.ok(m.decimalStringError(9007199254740993, 'id').includes('浮点'));
  assert.ok(m.decimalStringError(-1, 'id').includes('浮点'));
  assert.ok(m.decimalStringError('1.5', 'id').includes('纯十进制'));
  assert.ok(m.decimalStringError('1e10', 'id').includes('纯十进制'));
  assert.ok(m.decimalStringError('abc', 'id').includes('纯十进制'));
});

test('undefined/null/empty pass decimal check (missing stays missing)', () => {
  assert.equal(m.decimalStringError(undefined, 'id'), '');
  assert.equal(m.decimalStringError(null, 'id'), '');
});

test('checkDecimalFields checks multiple fields', () => {
  const errs = m.checkDecimalFields({ a: '123', b: 42.5, c: 'abc' }, ['a', 'b', 'c']);
  assert.equal(errs.length, 2);
});

test('unknown time is not forced to today/epoch', () => {
  assert.equal(m.timeUnknownError(null), '');
  assert.equal(m.timeUnknownError('unknown'), '');
  assert.ok(m.timeUnknownError('').includes('未知时间'));
});

test('ownerReport flags owners not registered in data-ownership.json', () => {
  const known = ['Identity/Policy', 'Media', 'Audit'];
  const r = m.ownerReport([{ owner: 'Media' }, { owner: 'NotRegistered' }, { owner: 'Media' }], known);
  assert.equal(r.total, 3);
  assert.deepEqual(r.unknownOwners, ['NotRegistered']);
});