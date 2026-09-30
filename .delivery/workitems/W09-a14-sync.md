# W09 串行共享窗口：A14 契约-文档一致性复核与同步

状态: 完成（只读复核 + 纯文档新增，node/python 全量回归通过）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改 + 61 处未跟踪未触碰（本轮只改 1 个文档文件，其余只读）

## 改动
- 新建 `scripts/a14_contract_doc_check.py`
  - 复用 W03/W04 结构校验器、W02b 生成器幂等检查
  - 新增文档措辞关键词核对：5 个已冻结 schema 的关键语义必须出现在
    docs/architecture 与 docs/operations 相关文档中（防文档与契约分叉）
  - 输出 `.delivery/reports/` 复核报告，退出码 0=一致 / 1=不一致
- 修改 `docs/architecture/架构与数据流.md`（纯文档新增，未碰代码）
  - 在"三个进度与两个授权域"节补充已冻结 schema 引用：
    media-receipt（耐久后 ACK/不假 ACK/backup_set）、media-manifest（object_ref
    复合键/sha256 不覆盖/quarantined）、integrity-report（counts 一致/三态）、
    license-claims（feature_grants 唯一来源/feature_catalog_version/license_revision/未签名非激活码）

## 证据
- `python3 scripts/a14_contract_doc_check.py` → 初跑 3 个关键词缺失（integrity-report/
  license-claims/media-manifest），补齐文档后 → 全部通过，exit=0
- 回归全绿：node 73/73、Python 4 校验器全绿（feature-catalog、contract-schemas、
  protocol-schemas、gen_feature_types --check）
- 用户 11 处已修改 + 61 处未跟踪未触碰

## 没有做/限制（如实）
- 本复核只做结构/关键词/同步检查，不替代 JWS 签名、时间单位、身份绑定、防回滚与
  配额语义校验（R42.8，需真实签发代码）
- 未更新 MANIFEST.sha256（保持既有范围差异口径）
- 未触碰 docs/operations 等其它文档（其措辞与契约已一致，无需补）

## 下一项
- 串行窗口可安全推进的纯新增/只读工作已到边界：合同线（W02-W08）+ 文档同步（W09）
  全部收口，73 测试全绿
- P1 跨层真实编码（server presence / Rust heartbeat / ingestion）仍等用户处理
  61 处未提交内容解锁 worktree 并行，或授权直接在主工作树写跨层代码