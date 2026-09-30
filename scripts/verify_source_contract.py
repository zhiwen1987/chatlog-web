#!/usr/bin/env python3
"""Offline structural checks for packages/source-contract/source-adapter.json (W14).

Enforces structural constraints only (no network, no product access):
- source-adapter: required top-level fields, capabilities sub-fields,
  identity/cursor/overwrite_policy required fields, watermark decimal-string shape
See 开发规则.md R05/R42.7 (signature/time/unit/identity checks are NOT here).
"""
from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SRC = ROOT / "packages" / "source-contract"


def load(name: str) -> dict:
    path = SRC / name
    with path.open(encoding="utf-8") as fh:
        data = json.load(fh)  # raises on invalid JSON
    if not isinstance(data, dict):
        raise ValueError(f"{name} must be a JSON object")
    return data


def check() -> list[str]:
    errors: list[str] = []
    sa = load("source-adapter.json")
    req = ("contract_version", "source_id", "capabilities", "identity", "cursor", "unsupported", "overwrite_policy")
    for f in req:
        if f not in sa.get("properties", {}):
            errors.append(f"source-adapter: missing property {f}")
    for f in sa.get("required", []):
        if f not in sa.get("properties", {}):
            errors.append(f"source-adapter: required field {f} has no property definition")

    caps = sa.get("properties", {}).get("capabilities", {}).get("required", [])
    for f in ("detect", "schema_probe", "stable_id", "shard_generation", "cursor"):
        if f not in caps:
            errors.append(f"source-adapter: capabilities missing {f}")

    ident = sa.get("properties", {}).get("identity", {}).get("required", [])
    for f in ("kind", "external_id"):
        if f not in ident:
            errors.append(f"source-adapter: identity missing {f}")

    cur = sa.get("properties", {}).get("cursor", {}).get("required", [])
    for f in ("watermark", "is_exhaustive"):
        if f not in cur:
            errors.append(f"source-adapter: cursor missing {f}")

    ow = sa.get("properties", {}).get("overwrite_policy", {}).get("required", [])
    for f in ("low_water_backfill", "duplicate_observation", "account_switch"):
        if f not in ow:
            errors.append(f"source-adapter: overwrite_policy missing {f}")
    return errors


def _validate_example(path: Path) -> list[str]:
    errors: list[str] = []
    data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        return [f"example {path.name}: must be a JSON object"]
    if "cursor" in data:
        wm = data.get("cursor", {}).get("watermark")
        if not isinstance(wm, str) or not wm.isdigit():
            errors.append(f"example {path.name}: watermark must be decimal string")
    if "capabilities" in data:
        for k, v in data.get("capabilities", {}).items():
            if k in ("detect", "schema_probe", "stable_id", "shard_generation", "cursor") and not isinstance(v, bool):
                errors.append(f"example {path.name}: capabilities.{k} must be boolean")
    return errors


def check_examples() -> tuple[int, list[str]]:
    ex = SRC / "examples"
    if not ex.is_dir():
        return 0, []
    errors: list[str] = []
    count = 0
    for path in sorted(ex.glob("*.json")):
        try:
            errors.extend(_validate_example(path))
        except (json.JSONDecodeError, OSError, ValueError) as exc:
            errors.append(f"example {path.name}: {exc}")
        count += 1
    return count, errors


def main() -> int:
    try:
        errors = check()
        n_ex, ex_errors = check_examples()
    except (OSError, ValueError, json.JSONDecodeError) as exc:
        print(f"ERROR: {exc}")
        return 1
    errors += ex_errors
    if errors:
        for err in errors:
            print(f"ERROR: {err}")
        print(f"FAIL {len(errors)} structural error(s)")
        return 1
    print(f"OK source-contract + {n_ex} example(s) pass structural checks")
    return 0


if __name__ == "__main__":
    sys.exit(main())