#!/usr/bin/env python3
"""Offline structural checks for packages/contracts/feature-catalog.json.

Enforces rules R24.1/R42.2 structural constraints only:
- unique stable feature keys (no all=true / wildcard / parent-prefix auto-grant)
- dependencies reference known keys (no dangling grant)
- dependency graph is a DAG (no cycle)
- each grant window is left-closed right-open (valid_from < valid_until)
No network, no product access, no signature verification (see 开发规则.md R42).
"""
from __future__ import annotations

import argparse
import json
import sys
from collections import defaultdict, deque
from pathlib import Path


def load(path: Path) -> dict:
    with path.open(encoding='utf-8') as fh:
        data = json.load(fh)  # raises on invalid JSON
    if not isinstance(data, dict) or 'features' not in data:
        raise ValueError('feature-catalog.json must be an object with a "features" array')
    return data


def check(data: dict) -> list[str]:
    errors: list[str] = []
    features = data['features']
    if not isinstance(features, list):
        errors.append('"features" must be an array')
        return errors

    keys: list[str] = []
    for idx, item in enumerate(features):
        if not isinstance(item, dict) or 'key' not in item:
            errors.append(f'features[{idx}]: missing "key"')
            continue
        key = item['key']
        if not isinstance(key, str) or not key.strip():
            errors.append(f'features[{idx}]: "key" must be a non-empty string')
            continue
        if key != key.strip() or any(ch.isspace() for ch in key):
            errors.append(f'features[{idx}]: key {key!r} must not contain whitespace')
            continue
        if key == 'all' or '*' in key:
            errors.append(f'features[{idx}]: wildcard/all key forbidden: {key!r}')
        keys.append(key)
        deps = item.get('depends_on', [])
        if not isinstance(deps, list) or not all(isinstance(d, str) for d in deps):
            errors.append(f'features[{idx}] {key}: depends_on must be an array of strings')

    seen: set[str] = set()
    for key in keys:
        if key in seen:
            errors.append(f'duplicate feature key: {key}')
        seen.add(key)

    # Dangling dependency references.
    key_set = set(keys)
    for idx, item in enumerate(features):
        key = item.get('key', '')
        for dep in item.get('depends_on', []):
            if dep not in key_set:
                errors.append(f'features[{idx}] {key}: depends on unknown key {dep!r}')

    # DAG via Kahn's algorithm.
    adj: dict[str, list[str]] = defaultdict(list)
    indeg: dict[str, int] = defaultdict(int)
    for item in features:
        key = item['key']
        for dep in item.get('depends_on', []):
            adj[dep].append(key)
            indeg[key] += 1
    queue = deque(k for k in key_set if indeg[k] == 0)
    order: list[str] = []
    while queue:
        node = queue.popleft()
        order.append(node)
        for nxt in adj[node]:
            indeg[nxt] -= 1
            if indeg[nxt] == 0:
                queue.append(nxt)
    if len(order) != len(key_set):
        cycle = sorted(k for k in key_set if indeg[k] > 0)
        errors.append(f'dependency cycle detected among keys: {cycle}')

    # Grant window semantics (left-closed, right-open).
    for idx, item in enumerate(features):
        key = item.get('key', '')
        valid_from = item.get('valid_from')
        valid_until = item.get('valid_until')
        if valid_from is not None and valid_until is not None:
            if valid_from >= valid_until:
                errors.append(f'features[{idx}] {key}: valid_from must be < valid_until')
        if item.get('quotas') is not None and not isinstance(item['quotas'], dict):
            errors.append(f'features[{idx}] {key}: quotas must be an object')

    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('path', nargs='?', default='packages/contracts/feature-catalog.json')
    args = parser.parse_args()
    path = Path(args.path)
    try:
        data = load(path)
        errors = check(data)
    except (OSError, ValueError, json.JSONDecodeError) as exc:
        print(f'ERROR: {path}: {exc}')
        return 1
    if errors:
        for err in errors:
            print(f'ERROR: {err}')
        print(f'FAIL {path}: {len(errors)} structural error(s)')
        return 1
    print(f'OK {path}: {len(data["features"])} features, unique keys, acyclic DAG')
    return 0


if __name__ == '__main__':
    sys.exit(main())