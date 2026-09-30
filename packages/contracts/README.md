# packages/contracts

唯一跨语言 wire/schema/事件合同源（规则 R05 L273）。
变更须进入串行共享窗口（手册 M08.4），单 owner 归口修改。

当前内容：
- `feature-catalog.json`：功能目录规范源（规则 R42.2/R24.1）。
  23 个功能 key、7 条依赖、DAG 无环。生成 Go/TS 类型、平台表单、
  产品状态映射与测试的唯一来源，禁止在 app 目录维护手写副本。
  校验入口：`scripts/verify_feature_catalog.py`（正例 + 坏例）。
- `data-ownership.json`：数据所有权字典（规则 R07/R39/R42 数据入口）。
  11 个 Owner 域 + 必须字段模式（64位ID/bytes/seq/revision 十进制字符串、
  tenant 复合 FK、分区唯一）。只登记 owner→边界与输入/输出，不复制 DDL。
- `license-claims.json`：商业许可 claims v2 骨架（规则 R42.8）。
  `feature_grants[]` 为商业权利唯一来源，替代旧 `features[]`；外层
  licensee/deployment/公钥绑定/license_revision/activation_epoch/lease_seq/
  mode/version_range/feature_catalog_version/全局 limits。未签名，不能作激活码。
- `integrity-report.json`：媒体完整性报告骨架（规则 R42.10）。
  原件状态三态 `excluded_by_license`/`paused_by_policy`/`source_missing`，
  不能全写 missing 更不能标 verified；counts 需一致（verified+pending+
  excluded+source_missing ≤ in_scope ≤ total_discovered）。
- 媒体合同已冻结在 `packages/protocol/`（R42.7，禁止在 app 目录重复定义）：
  - `media-manifest.json`：媒体对象清单（object_ref/bytes_ref/sha256/state）。
  - `media-receipt.json`：采集收据（仅在耐久提交后发出；receipt≠发布/索引/校验）。
  - 协议 schema 仍待冻结（packages/protocol 同步协议类型，R05 范围）。
- 来源契约已冻结在 `packages/source-contract/`（R05/R42.7，SourceAdapter 接口）：
  - `source-adapter.json`：detect/schema 探测/账号身份/shard generation/稳定 ID/
    游标能力/不支持情况/overwrite_policy；来源只读、毒数据隔离、不假 ACK。
  - 校验入口 `scripts/verify_source_contract.py`（正例 + 坏例）。
- AI 画像契约已冻结在 `packages/profile-schema/`（R05/R06-R10）：
  - `profile-schema.json`：identity/facts/evidence/snapshots 分层；画像是可重建
    派生非唯一数据源（derived.unique_source=false）、受限 Worker 生成、人工新版本
    不被 AI 覆盖。
  - 校验入口 `scripts/verify_profile_schema.py`（正例 + 坏例）。

校验入口：
- `scripts/verify_feature_catalog.py`（正例 + 坏例）
- `scripts/verify_contract_schemas.py`（W03 三个合同 JSON 结构校验 + 坏例）
- `scripts/verify_protocol_schemas.py`（W04 media-manifest/media-receipt 结构校验 + 坏例）
- `scripts/verify_source_contract.py`（W14 source-adapter 结构校验 + 坏例）

owner: 协调者（Lead）串行窗口；子代理只读建议。
