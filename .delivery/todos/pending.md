# 待办/待提醒清单

## P0 · 用户未提交内容处置 — ✅ 已解决（2026-09-30 授权提交基线）
- 用户已授权提交：63 处未提交/未跟踪内容已提交为基线（main 578cc8e，见 HANDOFF）
- worktree 隔离已解锁（wt-a/b/c 各 87 全绿）

## P1 · 真实业务编码
- ✅ W29-W31 心跳纵切已完成（server presence / Rust heartbeat / admin-web 设备页）
  （见 workitems/W29-device-heartbeat.md、W30-device-presence-page.md、W31-desktop-heartbeat.md）
- 🔄 W03 数据链（ingestion / upload / oracle）— 进行中（当前轮目标）

## P1.5 · 合同扩展（串行窗口，已全部收口）
- ✅ W02 feature-catalog 骨架（已完成，见 workitems/W02-feature-catalog.md）
- ✅ W03 合同扩展（已完成，见 workitems/W03-contracts-ext.md）：data-ownership / license claims v2 / IntegrityReport 骨架 + 校验器 + 反证测试
- ✅ W02b 类型生成器（已完成，见 workitems/W02b-gen-feature-types.md）：feature-catalog.json → Go 常量/依赖表 + JS FEATURES/FEATURE_CATALOG + --check 幂等守护 + 反证
- ✅ W04 媒体 manifest/收据协议 schema（已完成，见 workitems/W04-media-protocol.md）：packages/protocol/{media-manifest,media-receipt}.json + verify_protocol_schemas.py + 反证测试
- ✅ W05 integrity-report 前端消费工具（已完成，见 workitems/W05-integrity-tool.md）：admin-web src/lib/integrity.js + tests/integrity.test.mjs（node:test 8 个通过，回归 53 全绿）
- ✅ W06 license-claims 前端消费工具（已完成，见 workitems/W06-license-tool.md）：admin-web src/lib/license.js + tests/license.test.mjs（node:test 6 个通过，回归 59 全绿；含 grantActive/dependencyCheck/renderClaims）
- ✅ W07 media manifest/receipt 前端校验工具（已完成，见 workitems/W07-media-tool.md）：admin-web src/lib/media.js + tests/media.test.mjs（node:test 8 个通过，回归 67 全绿；含 manifestErrors/receiptErrors/reconcile 拒假 ACK）
- ✅ W08 data-ownership 前端校验工具（已完成，见 workitems/W08-ownership-tool.md）：admin-web src/lib/ownership.js + tests/ownership.test.mjs（node:test 6 个通过，回归 73 全绿；含 decimalStringError/checkDecimalFields/timeUnknownError/ownerReport）
- ✅ W09 A14 契约-文档一致性复核与同步（已完成，见 workitems/W09-a14-sync.md）：scripts/a14_contract_doc_check.py + docs/architecture/架构与数据流.md 补引用；3 个关键词缺失已修复，全部通过 exit=0
- ✅ W10 校验器加固 required 一致性（已完成，见 workitems/W10-validator-hardening.md）：反证发现 required 缺失抓不到，修 verify_contract_schemas/verify_protocol_schemas 的 4 个 schema 检查；坏例 exit=1、恢复后全绿

## P2 · 后续提醒点
- ✅ 未提交内容已处理（见 P0）
- ✅ 插件状态：用户已确认装好
