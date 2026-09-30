#!/usr/bin/env python3
"""Offline structural checks for packages/protocol/*.json (W04).

Enforces structural constraints only (no network, no signature verification):
- every file is a valid JSON object with required top-level fields
- media-manifest: sha256 hex(64), size non-negative, state enum,
  object_ref/bytes_ref present, origin required fields
- media-receipt: seq decimal-string shape, committed_at date-time,
  media_kind enum, object_ref/media_kind consistency
See 开发规则.md R42.7 (signature/time/unit/identity checks are NOT here).
"""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PROTOCOL = ROOT / "packages" / "protocol"


def load(name: str) -> dict:
    path = PROTOCOL / name
    with path.open(encoding="utf-8") as fh:
        data = json.load(fh)  # raises on invalid JSON
    if not isinstance(data, dict):
        raise ValueError(f"{name} must be a JSON object")
    return data


def check() -> list[str]:
    errors: list[str] = []

    # ---- media-manifest ----
    mm = load("media-manifest.json")
    req = ("manifest_id", "catalog_version", "tenant_id", "deployment_id",
           "object_ref", "media_type", "origin", "bytes_ref", "sha256",
           "size_bytes", "seq", "created_at", "state")
    for f in req:
        if f not in mm.get("properties", {}):
            errors.append(f"media-manifest: missing required field {f}")
    for f in mm.get("required", []):
        if f not in mm.get("properties", {}):
            errors.append(f"media-manifest: required field {f} has no property definition")
    if set(mm.get("required", [])) != set(req):
        errors.append("media-manifest: required list does not match the declared mandatory fields")
    sha = mm.get("properties", {}).get("sha256", {})
    if sha.get("pattern") not in ("^[a-f0-9]{64}$",):
        errors.append("media-manifest: sha256 must be hex(64) pattern")
    mt = mm.get("properties", {}).get("media_type", {}).get("enum", [])
    if set(mt) != {"image", "video", "audio", "file"}:
        errors.append("media-manifest: media_type enum must be the four media kinds")
    st = mm.get("properties", {}).get("state", {}).get("enum", [])
    if not set(st) <= {"stored", "deleting", "deleted", "quarantined"}:
        errors.append("media-manifest: state enum out of range")

    # ---- media-receipt ----
    mr = load("media-receipt.json")
    req = ("receipt_id", "catalog_version", "tenant_id", "deployment_id",
           "source", "source_message_ref", "seq", "committed_at", "backup_set")
    for f in req:
        if f not in mr.get("properties", {}):
            errors.append(f"media-receipt: missing required field {f}")
    for f in mr.get("required", []):
        if f not in mr.get("properties", {}):
            errors.append(f"media-receipt: required field {f} has no property definition")
    if set(mr.get("required", [])) != set(req):
        errors.append("media-receipt: required list does not match the declared mandatory fields")
    mk = mr.get("properties", {}).get("media_kind", {}).get("enum", [])
    if set(mk) != {"image", "video", "audio", "file"}:
        errors.append("media-receipt: media_kind enum must be the four media kinds")
    return errors


def _validate_example(path: Path) -> list[str]:
    errors: list[str] = []
    data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        return [f"example {path.name}: must be a JSON object"]

    if "sha256" in data:
        if not re.fullmatch(r"[a-f0-9]{64}", data["sha256"]):
            errors.append(f"example {path.name}: sha256 not hex(64)")
        if not isinstance(data.get("size_bytes"), int) or data["size_bytes"] < 0:
            errors.append(f"example {path.name}: size_bytes must be non-negative int")
        if data.get("state") not in ("stored", "deleting", "deleted", "quarantined"):
            errors.append(f"example {path.name}: bad state {data.get('state')!r}")
        origin = data.get("origin", {})
        for f in ("kind", "source", "source_message_ref"):
            if f not in origin:
                errors.append(f"example {path.name}: origin missing {f}")

    if "committed_at" in data:
        seq = data.get("seq")
        if not isinstance(seq, str) or not seq.isdigit():
            errors.append(f"example {path.name}: seq must be decimal string")
        mk = data.get("media_kind")
        if mk is not None and mk not in ("image", "video", "audio", "file"):
            errors.append(f"example {path.name}: bad media_kind {mk!r}")
        if (data.get("object_ref") is None) != (mk is None):
            errors.append(f"example {path.name}: object_ref and media_kind must both be set or both null")

    return errors


def check_examples() -> tuple[int, list[str]]:
    ex = PROTOCOL / "examples"
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
    print(f"OK protocol + {n_ex} example(s) pass structural checks")
    return 0


if __name__ == "__main__":
    sys.exit(main())