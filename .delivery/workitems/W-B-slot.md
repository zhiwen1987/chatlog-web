# W-B 槽 · A14 契约-文档一致性补缺（source-contract / profile-schema）

状态: 完成（纯文档+校验脚本扩展，回归全绿）
owner: B 槽执行者（wt-b-docs）｜分支 wt-b-docs｜worktree chatlog-web-wt-b
基线与任务: 核对 W02-W22 已冻结契约与文档一致性，补 A14 缺失项（不碰 schema/apps/router）

## 基线（先验证，均绿）
- `python3 scripts/verify_serial_window.py` → OK serial window，exit=0
- `python3 scripts/a14_contract_doc_check.py` → 初跑 exit=0（旧检查范围未覆盖 source/profile）

## 核对结论（docs 四目录 vs packages schema）
用 grep 核 key 词：feature_grants/object_ref/quarantined/receipt/backup_set/十进制字符串 均已引用；
发现缺失：`watermark`、`derived`、`unique_source` 在 docs/architecture|operations|install|mcp 零命中。
这两份 schema 是 4.3.1 已冻结契约（W14 source-contract、W15 profile-schema），但 a14 脚本
`CONTRACT_DOC_KEYWORDS` 未覆盖 → 确认 A14 缺口。

## 改动（4 文件，纯文档+脚本）
- `scripts/a14_contract_doc_check.py`：CONTRACT_DOC_KEYWORDS 增加
  - source-adapter.json: [watermark, 十进制字符串, 回放推进, 不假]
  - profile-schema.json: [derived, unique_source, 可重建派生, 唯一数据源]
  DOC_FILES 增加 docs/architecture/二次开发指南.md
- `docs/architecture/架构与数据流.md`：已冻结 schema 清单补 source-adapter（cursor.watermark
  单调 seq 十进制字符串回放推进/来源只读不假 ACK）与 profile-schema（derived.rebuilt/from/
  unique_source 必须 false，可重建派生非唯一数据源）两条
- `docs/architecture/模块与数据所有权.md`：Source/Sync 行补 cursor.watermark 字段引用；
  Customer/Profile 行补 derived.unique_source 必须 false + 可重建派生非唯一数据源
- `docs/architecture/二次开发指南.md`：扩展一补 游标 cursor.watermark 单调 seq 十进制字符串回放推进
- `CHANGELOG.md`：4.3.1 节追加「补 · A14 检查范围扩展（W-B 槽）」记录

## 反证（红→绿）
- 改脚本未补文档时 `a14_contract_doc_check.py` → exit=1：
  - source-adapter 缺失 ['watermark', '回放推进']
  - profile-schema 缺失 ['derived', 'unique_source']
- 补齐文档后 → 全部通过 exit=0

## 回归证据
- verify_serial_window.py → OK，exit=0
- a14_contract_doc_check.py → 全部通过，exit=0
- 6 个 Python 校验器全绿：verify_feature_catalog / verify_contract_schemas /
  verify_protocol_schemas / verify_source_contract / verify_profile_schema /
  gen_feature_types --check
- node 全量（apps/admin-web `npm test`）→ 87/87 通过
- git diff：仅 4 个文件，8 insertions(+) 3 deletions(-)，未触碰 packages/ schema、apps/ 代码、router

## 没有做/限制（如实）
- 未更新 MANIFEST.sha256 / docs/文档索引 checksum（既有口径：W09/W16 均不更新）
- 未触碰 packages/contracts、packages/protocol、packages/source-contract、packages/profile-schema
  的 schema 文件；未改 apps/admin-web 任何代码与 router
- 本复核仍只做结构/关键词/同步检查，不替代 JWS 签名、时间单位、身份绑定、防回滚与配额语义校验

## 下一项
- 报告协调者（lead）：A14 缺失项已补齐、校验器全绿、纯文档+脚本改动