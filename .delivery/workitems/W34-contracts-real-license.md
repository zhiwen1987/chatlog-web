# W34 前端接入真实 license status（ContractsView 接后端）

状态: 完成（node 94/94、lint 0 errors、vue build OK）
owner: 协调者（Lead）串行+全自动推进
base: main 54c77e8（工作树干净）

## 改动
- `apps/admin-web/src/api/enterprise.js`
  - 新增 `getLicenseStatus()`：调 `GET /api/v1/license/status`，对象响应原样透出
    `{present, mode, checked, licensee, deployment, features}`
  - 401/403 由 request 统一抛"拒绝访问"；5xx 抛 HTTP 状态错误
- `apps/admin-web/src/views/ContractsView.vue`
  - 新增"许可状态（服务端实时）"区块：Enterprise 模式且已登录才拉取
  - 真实数据优先：present/licensee/deployment/features（allowed/missing）卡片展示
  - 未接入/未登录/错误态显式展示，不冒充成功
  - 本地合成 claims 区块标注"本地演示对照"，不替代服务端判定
- `apps/admin-web/tests/enterprise.test.mjs`（新增）
  - 源码 alias 替换为内联 data: URL 桩后 import 真实方法（getLicenseStatus 不调用 data 模块）
  - 4 测试：200 对象解包 / 401 拒绝 / HTML 非数据拒绝 / 500 HTTP 错误

## 证据
- `node --test tests/enterprise.test.mjs`：4/4 PASS
- `npm test`：94/94 PASS（原 90 + 新 4）
- `npm run lint`：0 errors（14 存量 warnings 未新增）
- `npm run build`：DONE（dist 可部署）
- 提交 `510e573`，工作树干净

## 没有做/限制
- 未改服务端（W33 已提供 license/status，本次纯前端消费）
- 演示/未登录态仍展示本地合成 claims（标注演示，避免空白）；真实接入后以服务端为准
- 未接 ContractView 的 integrity-report / media 区块到后端（不在本工作项）

## 下一项
- 全量回归（go+node+python+desktop）确认无跨层回归
- 服务端镜像已上线（W33），前端 build 产物可随下个部署批次上线