# W20 串行共享窗口：integrity 伪造完整防护（verified 带 reason 拒绝）

状态: 完成（测试驱动发现真实语义缺口，修复+回归）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改文件未触碰

## 问题发现（测试驱动）
实测 `renderReport` 对 `items` 里 verified 项带 reason 的报告仍返回 ok=true——
`countErrors` 只查 counts 一致性，不查 items 语义。R42.10 明确：
verified 不可能带排除原因（不冒充完整）。

## 修复
- `apps/admin-web/src/lib/integrity.js` 的 `renderReport`
  - 检查 items：verified 项带 reason 即报"verified 项不能带 reason（不冒充完整）"
  - 与 counts 错误合并，抛 UNTRUSTED_COUNTS
- `apps/admin-web/tests/integrity.test.mjs` 补测试
  - verified 带 reason → 抛 UNTRUSTED_COUNTS；verified 不带 reason → 通过

## 证据
- 修复前实测：`verified 带 reason => ok true comp 1`（缺陷）
- 修复后实测：抛 UNTRUSTED_COUNTS（测试通过）
- `integrity.test.mjs` → 9/9；全量 `tests/*.test.mjs` → 82/82
- 5 个 Python 校验器 + A14 全绿
- 用户 11 处已修改文件未触碰

## 没有做/限制
- 未改 contracts/integrity-report.json（schema 的 items 约束已由 verify_contract_schemas
  的 example 断言覆盖；前端 renderReport 补齐消费端语义）
- 未更新 MANIFEST.sha256（保持既有口径）

## 下一项
- 串行窗口测试加固：前端 5 个工具边界全覆盖，后续可补 source-contract 前端消费
- P1 跨层真实编码仍等用户决策
