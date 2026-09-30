#!/usr/bin/env python3
"""Offline structural checks for packages/profile-schema/profile-schema.json (W15).

Enforces structural constraints only (no network, no product access):
- profile-schema: required top-level fields, layers enum subset,
  derived.unique_source must be false (derived, not primary source)
See 开发规则.md R05/R06-R10 (signature/time/unit/identity checks are NOT here).
"""
from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PROFILE = ROOT / "packages" / "profile-schema"
LAYERS = {"identity", "facts", "evidence", "snapshots"}


def load(name: str) -> dict:
    path = PROFILE / name
    with path.open(encoding="utf-8") as fh:
        data = json.load(fh)  # raises on invalid JSON
    if not isinstance(data, dict):
        raise ValueError(f"{name} must be a JSON object")
    return data


def check() -> list[str]:
    errors: list[str] = []
    ps = load("profile-schema.json")
    req = ("schema_version", "profile_id", "tenant_id", "subject_ref",
           "source", "layers", "generated_at", "derived")
    for f in req:
        if f not in ps.get("properties", {}):
            errors.append(f"profile-schema: missing property {f}")
    for f in ps.get("required", []):
        if f not in ps.get("properties", {}):
            errors.append(f"profile-schema: required field {f} has no property definition")

    layer_enum = ps.get("properties", {}).get("layers", {}).get("items", {}).get("enum", [])
    if not set(layer_enum) <= LAYERS:
        errors.append("profile-schema: layers enum out of range")

    derived = ps.get("properties", {}).get("derived", {}).get("properties", {})
    for f in ("rebuilt", "from", "unique_source"):
        if f not in derived:
            errors.append(f"profile-schema: derived missing {f}")
    return errors


def _validate_example(path: Path) -> list[str]:
    errors: list[str] = []
    data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        return [f"example {path.name}: must be a JSON object"]
    for l in data.get("layers", []):
        if l not in LAYERS:
            errors.append(f"example {path.name}: bad layer {l!r}")
    d = data.get("derived", {})
    if d.get("unique_source") is not False:
        errors.append(f"example {path.name}: derived.unique_source must be false (画像是可重建派生，非唯一数据源)")
    return errors


def check_examples() -> tuple[int, list[str]]:
    ex = PROFILE / "examples"
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
    print(f"OK profile-schema + {n_ex} example(s) pass structural checks")
    return 0


if __name__ == "__main__":
    sys.exit(main())