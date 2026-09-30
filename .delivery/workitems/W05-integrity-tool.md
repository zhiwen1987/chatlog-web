# W05 串行共享窗口：integrity-report 前端消费工具（admin-web src/lib）

状态: 完成（纯新增 JS + node:test 真实通过）
owner: 协调者（Lead）串行窗口
base: 用户 59 处未提交内容保留原样（本轮未触碰任何既有未提交文件）

## 改动（全部为只增）
- 新建 `apps/admin-web/src/lib/integrity.js`
  - `countErrors(counts)`：完整性报告 counts 一致性校验（R42.10）
    verified+pending+excluded+source_missing ≤ in_scope ≤ total_discovered；
    缺失/非整数拒绝，不补 0
  - `statusLabel(status)`：三态中文标签
    excluded_by_license=未授权排除 / paused_by_policy=策略暂停 / source_missing=源缺失
  - `renderReport(report)`：可信展示结构；counts 不一致抛
    `UNTRUSTED_COUNTS`（不显示假完整）；completeness=verified/in_scope，
    in_scope=0 时为 null 不编造；outOfScope=total-in_scope
- 新建 `apps/admin-web/tests/integrity.test.mjs`
  - 与现有 tests/core.test.mjs 同风格：`node:test` + data:text/javascript 动态加载

## 证据
- `node --test tests/integrity.test.mjs` → pass 8 / fail 0
- 回归：`node --test tests/core.test.mjs tests/rpc.test.mjs tests/sql-boundary.test.mjs`
  → pass 22 / fail 0；`tests/wcdb.test.mjs` → pass 23 / fail 0
- 合计 53 个测试全绿（新增 8 + 回归 45）
- 纯函数无副作用；只消费报告对象，不篡改数据

## 没有做/限制（如实）
- 未接 UI 视图（新工具未在 .vue 中挂载；Media.vue 等视图是用户已修改文件，不触碰）
- 未更新 MANIFEST.sha256（保持既有范围差异口径）
- 未做签名/时间/单位语义校验（R42.8，需真实签发代码）

## 下一项
- 串行纵切可继续：同模式新增前端消费工具（如 license-claims 前端展示、feature 依赖检查 UI 逻辑）
- P1 跨层真实编码（server presence / Rust heartbeat）仍等用户处理 59 处未提交内容解锁 worktree