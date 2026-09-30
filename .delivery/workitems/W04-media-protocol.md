# W04 串行共享窗口：媒体 manifest / 收据协议 schema（packages/protocol）

状态: 完成（骨架 + 结构校验 + 反证测试）
owner: 协调者（Lead）串行窗口
base: 用户 59 处未提交内容保留原样（仅新增文件，未触碰任何既有未提交内容）

## 改动（全部为只增）
- 新建 `packages/protocol/media-manifest.json`
  - 媒体对象清单（R42.7）：manifest_id/catalog_version/tenant_id/deployment_id/
    object_ref/media_type/origin(bytes_ref/sha256/size_bytes/seq/created_at/state)
  - origin.kind 区分 original/derived（派生件不覆盖原件，R42.10）
  - state 含 quarantined（毒数据隔离，不假 ACK）
  - notes: 不覆盖 hash / 不 UI 计数遮盖 / object_ref 复合分区唯一
- 新建 `packages/protocol/media-receipt.json`
  - 采集收据：receipt_id/catalog_version/tenant_id/deployment_id/source/
    source_message_ref/seq/committed_at/backup_set
  - notes 对齐既有文档语义：ACK 仅在耐久提交后发出；receipt 不代表已发布/
    已索引/媒体已验证；PG 损坏/丢失时不发假 ACK（备份恢复与应急）
  - object_ref/media_kind 可空且成对（非媒体消息收据不带对象）
- 新建 `scripts/verify_protocol_schemas.py`
  - 结构校验（正例）：JSON 合法、必填字段、sha256 hex(64)、media_type/media_kind
    四类枚举、state 枚举、示例一致性断言
- 新建 `packages/protocol/examples/media-manifest.valid.json` +
  `media-receipt.valid.json`
- 更新 `packages/contracts/README.md`（登记 media 合同已冻结到 protocol 包 + 校验入口）

## 证据
- `python3 scripts/verify_protocol_schemas.py`
  → OK protocol + 2 example(s) pass structural checks（exit=0）
- 坏例反证（临时改坏后已还原）：
  - sha256 非 hex → 拒绝
  - state='bogus' → 拒绝
  - origin 缺 source_message_ref → 拒绝
  - FAIL 3 structural error(s), exit=1；还原后恢复 OK
- 4 个新 JSON 全部通过 python3 json.load 语法验证

## 没有做/限制（如实）
- 未做 JWS 签名/时间/单位/身份绑定/防回滚/配额语义校验（R42.8，需真实签发代码）
- protocol 包仍无 package.json scripts（`@chatlog/protocol` 0.1.0 空包），
  校验入口在根 scripts/，未改依赖
- 未更新 MANIFEST.sha256（保持既有范围差异口径）
- media-receipt 未含 media_verified/backed_up（属架构与数据流的其它标记，
  与 receipt 分离，不在本 schema 混入）

## 下一项
- 合同线已全部收口（W02/W03/W02b/W04）。P1 真实业务编码（心跳/数据链）等待
  用户处理 59 处未提交内容解锁 worktree 并行
- 或串行只读复核 A14：行为变更后同步功能说明/接口文档（待业务编码后）