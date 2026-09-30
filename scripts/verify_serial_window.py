#!/usr/bin/env python3
"""串行窗口综合终检（W17，A14 收尾验收门禁）。

聚合验证：
1. 所有合同 schema 文件存在（packages/contracts|protocol|source-contract|profile-schema）
2. 所有校验器通过（feature/contract/protocol/source/profile + gen --check）
3. 前端工具 + 测试文件存在（integrity/license/media/ownership/contracts）
4. 合同 schema 在 docs 中被引用（A14 关键词核对，复用 a14_contract_doc_check）
5. 前端全量测试入口存在（tests/*.test.mjs 通配）

只读 + 调用既有校验器；不修改任何源。退出码 0=全部通过。
"""
from __future__ import annotations

import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

SCHEMAS = [
    "packages/contracts/feature-catalog.json",
    "packages/contracts/data-ownership.json",
    "packages/contracts/license-claims.json",
    "packages/contracts/integrity-report.json",
    "packages/protocol/media-manifest.json",
    "packages/protocol/media-receipt.json",
    "packages/source-contract/source-adapter.json",
    "packages/profile-schema/profile-schema.json",
]

VALIDATORS = [
    "scripts/verify_feature_catalog.py",
    "scripts/verify_contract_schemas.py",
    "scripts/verify_protocol_schemas.py",
    "scripts/verify_source_contract.py",
    "scripts/verify_profile_schema.py",
]

FRONTEND_LIBS = [
    "apps/admin-web/src/lib/integrity.js",
    "apps/admin-web/src/lib/license.js",
    "apps/admin-web/src/lib/media.js",
    "apps/admin-web/src/lib/ownership.js",
    "apps/admin-web/src/lib/contracts.js",
    "apps/admin-web/src/lib/source.js",
]

FRONTEND_TESTS = [
    "apps/admin-web/tests/integrity.test.mjs",
    "apps/admin-web/tests/license.test.mjs",
    "apps/admin-web/tests/media.test.mjs",
    "apps/admin-web/tests/ownership.test.mjs",
    "apps/admin-web/tests/contracts.test.mjs",
    "apps/admin-web/tests/source.test.mjs",
]


def run(cmd: list[str]) -> tuple[int, str]:
    proc = subprocess.run(cmd, cwd=ROOT, capture_output=True, text=True)
    return proc.returncode, (proc.stdout + proc.stderr).strip()


def main() -> int:
    errors: list[str] = []
    for rel in SCHEMAS:
        if not (ROOT / rel).is_file():
            errors.append(f"缺少 schema: {rel}")
    for rel in FRONTEND_LIBS + FRONTEND_TESTS:
        if not (ROOT / rel).is_file():
            errors.append(f"缺少前端产物: {rel}")
    for script in VALIDATORS:
        code, out = run(["python3", script])
        if code != 0:
            errors.append(f"{script} 失败：{out}")
    code, out = run(["python3", "scripts/gen_feature_types.py", "--check"])
    if code != 0:
        errors.append(f"生成器失步：{out}")
    code, _ = run(["python3", "scripts/a14_contract_doc_check.py"])
    if code != 0:
        errors.append("A14 契约-文档复核未通过")
    code, out = subprocess.run(["node", "--test", "tests/*.test.mjs"],
                                cwd=ROOT / "apps/admin-web", capture_output=True, text=True).returncode and (1, ""), ""
    if code == 0:
        proc = subprocess.run(["node", "--test", "tests/*.test.mjs"],
                              cwd=ROOT / "apps/admin-web", capture_output=True, text=True)
        out = proc.stdout + proc.stderr
        if proc.returncode != 0:
            errors.append(f"前端测试入口执行失败：{out[:200]}")
        else:
            for line in out.splitlines():
                if line.startswith("ℹ fail") and "0" not in line:
                    errors.append(f"前端测试有失败：{line}")
                    break
    if code == 0:
        tail = out.splitlines()
        for line in tail:
            if line.startswith("ℹ fail") and "0" not in line:
                errors.append(f"前端测试有失败：{line}")
                break
    else:
        errors.append(f"前端测试入口执行失败：{out[:200]}")

    if errors:
        for e in errors:
            print(f"ERROR: {e}")
        print(f"FAIL {len(errors)} serial-window check(s)")
        return 1
    print("OK serial window: schemas present, validators pass, frontend tools+protocols present")
    return 0


if __name__ == "__main__":
    sys.exit(main())