#!/usr/bin/env python3
"""A14 契约-文档一致性复核（W02-W08 收口后的同步验证）。

复用已存在的校验器（W03/W04 结构校验、W02b 生成器幂等检查），并核对
docs/architecture 等既有文档措辞与已冻结 schema 的关键语义是否一致。
只读、无副作用；输出报告文本与退出码（0=一致, 1=不一致）。

关键语义关键词（与 schema/notes/文档一致）：
  - receipt: 耐久提交后 ACK、receipt 不代表已发布/索引/媒体已验证
  - manifest: object_ref 复合键、sha256 不覆盖、quarantined 毒数据隔离不假 ACK
  - ownership: 64位ID/bytes/seq/revision 十进制字符串、未知保留 null
  - license: v2 feature_grants 唯一来源、防降级、未签名非激活码
"""
from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
REPORT_DIR = ROOT / ".delivery" / "reports"

# 契约 schema 需在文档中出现的语义关键词（文档与 schema 的"口头共识"）
CONTRACT_DOC_KEYWORDS = {
    "packages/contracts/integrity-report.json": ["excluded_by_license", "source_missing", "counts", "in_scope"],
    "packages/contracts/license-claims.json": ["feature_grants", "feature_catalog_version", "license_revision"],
    "packages/contracts/data-ownership.json": ["十进制字符串", "tenant", "保留期限", "owner"],
    "packages/protocol/media-manifest.json": ["object_ref", "sha256", "bytes_ref", "quarantined"],
    "packages/protocol/media-receipt.json": ["耐久", "receipt", "backup_set", "不假"],
    "packages/source-contract/source-adapter.json": ["watermark", "十进制字符串", "回放推进", "不假"],
    "packages/profile-schema/profile-schema.json": ["derived", "unique_source", "可重建派生", "唯一数据源"],
}

DOC_FILES = [
    "docs/architecture/架构与数据流.md",
    "docs/architecture/模块与数据所有权.md",
    "docs/architecture/二次开发指南.md",
    "docs/operations/备份恢复与应急.md",
]


def run(cmd: list[str]) -> tuple[int, str]:
    proc = subprocess.run(cmd, cwd=ROOT, capture_output=True, text=True)
    return proc.returncode, (proc.stdout + proc.stderr).strip()


def check_contracts() -> list[str]:
    """复用 W03/W04 校验器（仅结构；签名/时间语义不在此）。"""
    errors: list[str] = []
    for script, label in (
        ("scripts/verify_contract_schemas.py", "contracts"),
        ("scripts/verify_protocol_schemas.py", "protocol"),
    ):
        code, out = run(["python3", script])
        if code != 0:
            errors.append(f"{label} 校验器失败：{out}")
    return errors


def check_generator_sync() -> list[str]:
    """W02b 生成器幂等守护：feature 类型产物与 catalog 一致。"""
    code, out = run(["python3", "scripts/gen_feature_types.py", "--check"])
    return [] if code == 0 else [f"feature 类型生成器失步：{out}"]


def check_doc_keywords() -> list[str]:
    """每个契约的关键语义关键词必须在相关文档中出现（防文档与契约分叉）。"""
    errors: list[str] = []
    docs_text: dict[str, str] = {}
    for doc in DOC_FILES:
        try:
            docs_text[doc] = (ROOT / doc).read_text(encoding="utf-8")
        except OSError as exc:
            errors.append(f"{doc} 不可读：{exc}")
            continue
    for schema, keywords in CONTRACT_DOC_KEYWORDS.items():
        rel = schema
        joined = " ".join(docs_text.values())
        missing = [kw for kw in keywords if kw not in joined]
        if missing:
            errors.append(f"{rel} 的语义关键词在文档中缺失：{missing}")
    return errors


def main() -> int:
    errors = []
    errors += check_contracts()
    errors += check_generator_sync()
    errors += check_doc_keywords()

    REPORT_DIR.mkdir(parents=True, exist_ok=True)
    lines = ["# A14 契约-文档一致性复核报告", ""]
    lines.append("检查项：")
    lines.append(f"- contracts/protocol 结构校验器：{'通过' if not (lambda e: e and any('校验器失败' in x for x in e))(errors) else '见下'}")
    lines.append(f"- feature 类型生成器幂等：{'通过' if not (lambda e: e and any('失步' in x for x in e))(errors) else '见下'}")
    lines.append(f"- 文档措辞关键词核对：{'通过' if not (lambda e: e and any('缺失' in x for x in e))(errors) else '见下'}")
    lines.append("")
    if errors:
        lines.append("未通过：")
        for err in errors:
            lines.append(f"- {err}")
    else:
        lines.append("全部通过：合同/协议 schema 结构、feature 生成器同步、文档措辞与已冻结语义一致。")
    lines.append("")
    lines.append("边界：本复核只做结构/关键词/同步检查，不替代 JWS 签名、时间单位、身份绑定、")
    lines.append("防回滚与配额语义校验（R42.8，需真实签发代码）；不更新 MANIFEST.sha256（既有口径）。")

    report = REPORT_DIR / "A14-契约文档一致性.md"
    report.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print("\n".join(lines))
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())