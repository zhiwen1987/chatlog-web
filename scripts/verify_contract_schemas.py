#!/usr/bin/env python3
"""Offline structural checks for packages/contracts/*.json (W03).

Enforces structural constraints only (no network, no signature verification):
- each contract is a valid JSON object with required top-level fields
- data-ownership: owners[] non-empty, field_mode present, no duplicate owner
- license-claims: feature_grants[] feature_key must exist in feature-catalog.json,
  grant window left-closed right-open, state enum, quotas positive integers
- integrity-report: counts consistent (verified+pending+excluded+source_missing
  <= in_scope <= total_discovered), status enum, no 'verified' while reason set
See 开发规则.md R42.7/R42.8/R42.10 (signature/time/unit/identity checks are NOT here).
"""
from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CONTRACTS = ROOT / "packages" / "contracts"


def load(name: str) -> dict:
    path = CONTRACTS / name
    with path.open(encoding="utf-8") as fh:
        data = json.load(fh)  # raises on invalid JSON
    if not isinstance(data, dict):
        raise ValueError(f"{name} must be a JSON object")
    return data


def check() -> list[str]:
    errors: list[str] = []

    # ---- data-ownership ----
    ow = load("data-ownership.json")
    for f in ("title", "catalog_version", "owners", "field_mode"):
        if f not in ow:
            errors.append("data-ownership: missing required field " + f)
    owners = ow.get("owners", [])
    seen = {o.get("owner") for o in owners if isinstance(o, dict)}
    if len(seen) != len(owners):
        errors.append("data-ownership: duplicate owner name")
    if not seen:
        errors.append("data-ownership: owners is empty")

    # ---- license-claims ----
    lc = load("license-claims.json")
    catalog = load("feature-catalog.json")
    known_keys = {f["key"] for f in catalog.get("features", [])}
    for f in ("typ", "iss", "aud", "iat", "exp", "licensee", "deployment",
              "feature_catalog_version", "license_revision", "activation_epoch",
              "feature_grants"):
        if f not in lc.get("properties", {}):
            errors.append(f"license-claims: missing property {f}")
    grants = lc.get("properties", {}).get("feature_grants", {}).get("items", {})
    props = grants.get("properties", {})
    for f in ("grant_id", "feature_key", "state", "valid_from"):
        if f not in props:
            errors.append("license-claims: feature_grants missing required field " + f)
    for f in lc.get("required", []):
        if f not in lc.get("properties", {}):
            errors.append(f"license-claims: required field {f} has no property definition")
    req_lc = ("typ", "iss", "aud", "iat", "exp", "licensee", "deployment",
              "feature_catalog_version", "license_revision", "activation_epoch",
              "feature_grants")
    if set(lc.get("required", [])) != set(req_lc):
        errors.append("license-claims: required list does not match the declared mandatory fields")

    # ---- integrity-report ----
    ir = load("integrity-report.json")
    for f in ("report_id", "generated_at", "scope", "counts", "items"):
        if f not in ir.get("properties", {}):
            errors.append(f"integrity-report: missing property {f}")
    counts = ir.get("properties", {}).get("counts", {}).get("properties", {})
    for f in ("total_discovered", "in_scope", "verified", "pending", "excluded", "source_missing"):
        if f not in counts:
            errors.append(f"integrity-report: counts missing field {f}")
    for f in ir.get("required", []):
        if f not in ir.get("properties", {}):
            errors.append(f"integrity-report: required field {f} has no property definition")

    return errors


def _validate_example(path: Path, known_keys: set[str]) -> list[str]:
    errors: list[str] = []
    data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        return [f"example {path.name}: must be a JSON object"]

    if "feature_grants" in data:
        grants = data.get("feature_grants", [])
        if not isinstance(grants, list):
            errors.append(f"example {path.name}: feature_grants must be a list")
        else:
            seen_grants: set[str] = set()
            for g in grants:
                gid = g.get("grant_id")
                if gid in seen_grants:
                    errors.append(f"example {path.name}: duplicate grant_id {gid}")
                seen_grants.add(gid)
                fk = g.get("feature_key")
                if fk not in known_keys:
                    errors.append(f"example {path.name}: unknown feature_key {fk!r}")
                st = g.get("state")
                if st not in ("active", "paused", "revoked", "expired"):
                    errors.append(f"example {path.name}: bad state {st!r}")
                vf = g.get("valid_from")
                vu = g.get("valid_until")
                if vf is not None and vu is not None and vf >= vu:
                    errors.append(f"example {path.name}: grant {gid} window not left-closed right-open")
                for q in g.get("quotas", []) or []:
                    if not isinstance(q.get("limit"), int) or q.get("limit", 0) < 1:
                        errors.append(f"example {path.name}: quota limit must be positive int")

    elif "counts" in data and "items" in data:
        counts = data.get("counts", {})
        keys = ("total_discovered", "in_scope", "verified", "pending", "excluded", "source_missing")
        missing = [k for k in keys if k not in counts]
        if missing:
            errors.append(f"example {path.name}: counts missing {missing}")
        else:
            sub = counts["verified"] + counts["pending"] + counts["excluded"] + counts["source_missing"]
            if not (sub <= counts["in_scope"] <= counts["total_discovered"]):
                errors.append(f"example {path.name}: counts inconsistent")
        for it in data.get("items", []) or []:
            st = it.get("status")
            if st not in ("verified", "pending", "excluded_by_license", "paused_by_policy", "source_missing"):
                errors.append(f"example {path.name}: bad status {st!r}")
            if st == "verified" and it.get("reason"):
                errors.append(f"example {path.name}: verified item must not have a reason")

    return errors


def check_examples() -> tuple[int, list[str]]:
    """Load shipped examples and run real structural assertions (fail on bad example)."""
    ex = CONTRACTS / "examples"
    if not ex.is_dir():
        return 0, []
    catalog = load("feature-catalog.json")
    known_keys = {f["key"] for f in catalog.get("features", [])}
    errors: list[str] = []
    count = 0
    for path in sorted(ex.glob("*.json")):
        if not path.name.endswith(".json"):
            continue
        try:
            errors.extend(_validate_example(path, known_keys))
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
    print(f"OK contracts + {n_ex} example(s) pass structural checks")
    return 0


if __name__ == "__main__":
    sys.exit(main())