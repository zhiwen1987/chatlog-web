# 待办/待提醒清单

## P0 · 用户未提交内容处置（阻塞三槽真实编码）
- 用户选择: 保持现状，串行或只读并行；后期提醒
- 内容: 59 处未提交/未跟踪（admin-web 图片/BASELINE、enterprise.go+test、desktop-client 全套、v4.3/、zip、docs 等）
- 影响: M08.2 worktree 隔离未满足 → 真实编码需主工作树串行或用户确认基线
- 下一步: 在用户确认前，只做只读复核/串行小项；此条保留到用户明确处理

## P1 · 真实业务编码（部分已串行推进，跨层仍等待基线）
- W02b 心跳纵切（server presence / Rust heartbeat / admin-web 设备页）
- W03 数据链（ingestion / upload / oracle）
- 需 worktree 或用户提交基线后方可三槽并行

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
- 提醒用户: 处理 59 处未提交内容以解锁三槽并行
- 提醒用户: 安装插件状态（用户已说装好）
