# W27 全自动推进：license status 实时查库（端到端验证）

状态: 完成（跨层 Go，真实 PG 端到端跑通）
owner: 协调者（Lead）全自动推进
base: main 806b9e7（W26 写端点+集成修复）

## 发现（端到端验证暴露）
写 license 返回 200，但 status 读启动时快照（s.License）→ present:false。
运行中写入的新 license 不生效（需重启才可见）——真实语义缺陷。

## 修复
- `apps/server/internal/handler/license.go`
  - status 端点实时查 DB：`db.LoadLicenseClaims(ctx, s.DB, s.DeploymentID)`
    （deployment 最新 revision）
  - DB 为 nil 时（单测/无 DB 环境）回退启动快照 s.License
- `apps/server/internal/handler/server.go`：Deps 加 `DeploymentID`
- `apps/server/cmd/server/main.go`：Deps 填 cfg.DeploymentID

## 证据（真实 PG 端到端，2026-09-30）
- 全量真实 PG：`go test -count=1 ./...` 全包 ok
- 端到端（临时实例 :8081 连真实 PG）：
  register 拿 token → POST /api/v1/license (deployment_id=dev) → 200
  → GET /api/v1/license/status → `present: True licensee: acme archive.read Allowed: True`
- 未认证写 → 401（走 Authenticate 中间件）
- node 90/90；综合终检 OK
- 调试修复：license.go 缺 model import；单测 fallback 保真（DB nil 用快照）

## 没有做/限制
- 未重建运行中 docker 容器（保留用户 dev 服务原样；新代码用临时实例验证）
- 未做 JWS 签名验证（需真实签发方）

## 下一步
- license 链端到端闭环完成：schema→model→auth→handler(读+写)→db→启动→实时查询
- 可继续：heartbeat/presence 端点纵切、或前端 ContractsView 接真实 /api/v1/license/status
