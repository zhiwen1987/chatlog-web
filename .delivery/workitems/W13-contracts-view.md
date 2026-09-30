# W13 串行共享窗口：契约与完整性 UI 视图（admin-web ContractsView.vue）

状态: 完成（纯新增组件 + 1 行路由，vue-cli build 编译通过）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改文件未触碰；router/index.js 为基线跟踪文件（未被用户改）安全修改

## 改动
- 新建 `apps/admin-web/src/views/ContractsView.vue`
  - 展示三个契约工具区块：完整性报告 counts 一致性、许可 claims v2 grant 状态、
    媒体收据/清单对账（拒假 ACK）
  - 消费 `@/lib/contracts` 统一入口（W11 的 validateContract/renderReport/renderClaims/reconcile）
  - 合成展示数据标注"真实运行时由服务端下发"，不替代签名/时间/身份语义
- 修改 `apps/admin-web/src/router/index.js`（+1 行）
  - 新增 `/contracts` 路由 → ContractsView，title "契约与完整性"

## 证据
- `npx vue-cli-service build` → DONE Build complete, exit=0（新组件+路由编译通过）
- dist 为 git 忽略产物，已清理无残留
- node 全量测试 79/79 通过（未影响既有前端测试）
- 用户 11 处已修改文件未触碰；router 是基线跟踪文件安全修改

## 没有做/限制（如实）
- 视图用合成展示数据，未接真实服务端下发（服务端 API 未实现，P0 阻塞）
- 未在导航菜单挂入口（Layout 菜单是用户已修改文件，不触碰；路由可直接访问 /contracts）
- 未做 JWS 签名/时间单位/身份绑定/防回滚/配额语义校验（R42.8，需真实签发代码）
- 未更新 MANIFEST.sha256（保持既有口径）

## 下一项
- 串行窗口：合同线 + 前端工具 + UI 视图已收口（W02-W13）
- P1 跨层真实编码（服务端 API/心跳/ingestion）仍等用户决策解锁 worktree 或授权直改
