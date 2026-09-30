# W03 串行共享窗口：合同扩展（data-ownership / license claims / IntegrityReport）

状态: 完成（骨架 + 结构校验 + 反证测试）
owner: 协调者（Lead）串行窗口
base: 未提交改动保留原样（用户 59 处未提交内容未触碰）

## 改动（全部为只增，未覆盖/删除现有文件）
- 新建 `packages/contracts/data-ownership.json`
  - 11 个 Owner 域（Identity/Policy、Device/Presence、Source/Sync、Chat/Query、
    Media、Customer/Profile、Jobs/Projection、Quality/Retention、
    Customer Entitlement、Issuer Licensing、Audit），每项登记 writes/io/retention
  - `field_mode`（64位ID/bytes/seq/revision 十进制字符串、tenant 复合 FK、
    分区唯一、时间字段分命名、id_kinds 区分）
  - `delivery_checklist`（每表 owner/migration/主键/FK/RLS/TTL；无 owner 不合并）
- 新建 `packages/contracts/license-claims.json`
  - claims v2 骨架（R42.8）：`feature_grants[]` 替代旧 `features[]`
  - 外层 licensee/deployment/public_key_binding/license_revision/activation_epoch/
    lease_seq/mode/version_range/feature_catalog_version/limits
  - feature_grants 每项含 grant_id/feature_key/state/valid_from/valid_until/quotas
  - 语义注记：v2-only 防降级、旧 features 迁移须发行方映射、未签名不作激活码
- 新建 `packages/contracts/integrity-report.json`
  - R42.10 三态：`excluded_by_license`/`paused_by_policy`/`source_missing`
  - counts 一致性约束（verified+pending+excluded+source_missing ≤ in_scope ≤ total_discovered）
  - scope 声明当前范围，不冒充整个微信 100%
- 新建 `scripts/verify_contract_schemas.py`
  - 结构校验（正例）：JSON 合法、必填字段、owner 唯一
  - 示例断言（反证）：feature_key 引用存在、state 枚举、grant 窗口左闭右开、
    quota 正整数、counts 一致、verified 不带 reason
- 新建 `packages/contracts/examples/license-claims.valid.json` + `integrity-report.valid.json`
- 更新 `packages/contracts/README.md`（登记三个新合同 + 校验入口）

## 证据
- `python3 scripts/verify_contract_schemas.py`
  → OK contracts + 2 example(s) pass structural checks（exit=0）
- 三条坏例反证（临时文件，验证后已清理）：
  - bad1 counts 不一致（verified=9>in_scope=8）→ 拒绝 `counts inconsistent`
  - bad2 verified 带 reason → 拒绝 `verified item must not have a reason`
  - bad3 unknown feature_key `no.such.key` → 拒绝 `unknown feature_key`
- 5 个合同 JSON 全部通过 python3 json.load 语法验证

## 没有做（待后续工作项）
- 未生成 Go/TS 类型、平台表单映射（W02b，R42.2）
- 未做 JWS 签名/时间/单位/依赖/身份绑定/防回滚/配额语义校验（R42.8，需真实签发代码）
- 未更新 MANIFEST.sha256（保持既有范围差异口径）
- 媒体 manifest/收据、协议 schema（packages/protocol）仍待冻结

## 下一项
- W02b 类型生成器：从 feature-catalog.json 生成 Go/TS 类型 + 平台表单映射
- 或 W04 媒体 manifest/收据合同