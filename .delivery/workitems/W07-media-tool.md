# W07 串行共享窗口：media manifest/receipt 前端校验工具（admin-web src/lib）

状态: 完成（纯新增 JS + node:test 真实通过）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改 + 43 处未跟踪未触碰（本轮未触碰任何既有文件）

## 改动（全部为只增）
- 新建 `apps/admin-web/src/lib/media.js`
  - `manifestErrors(m)`：媒体清单校验（R42.7）
    必填字段、sha256 hex(64)、media_type/state 枚举、origin.kind（original/derived）
  - `receiptErrors(r)`：采集收据校验
    必填字段、seq 十进制字符串、media_kind 枚举、object_ref/media_kind 成对出现
  - `reconcile(receipt, manifest)`：收据/清单对账
    - 非媒体收据（无对象）无需清单
    - 收据引用对象但无清单/清单非法/对象 quarantined（毒数据隔离）→ 拒绝，不发假 ACK
- 新建 `apps/admin-web/tests/media.test.mjs`
  - node:test + data:text/javascript 动态加载

## 证据
- `node --test tests/media.test.mjs` → pass 8 / fail 0
- 全量回归 `node --test tests/*.test.mjs` → tests 67 / pass 67 / fail 0
  （新增 media 8 + license 6 + integrity 8 + 既有 45）
- 对账反证：receipt 引用对象无 manifest → 拒绝假 ACK；quarantined 对象 → 拒绝

## 没有做/限制（如实）
- 未接 UI 视图（未在任何 .vue 挂载；Media 等视图是用户已修改文件，不触碰）
- 未做 JWS 签名验证/时间单位/身份绑定/防回滚/配额语义校验（R42.8，需真实签发代码）
- 未更新 MANIFEST.sha256（保持既有范围差异口径）
- media-manifest 的 source_message_ref 与 receipt 的 source_message_ref 关联
  仅在 reconcile 中按 object_ref 对账，未做跨表 join（协议层不负责）

## 下一项
- 串行纵切可继续：同模式新增前端消费工具（如 backup.schedule / quality.manage 授权展示）
- P1 跨层真实编码（server presence / Rust heartbeat / ingestion）仍等用户处理
  未提交内容解锁 worktree 并行