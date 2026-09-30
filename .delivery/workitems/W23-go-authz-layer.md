# W23 跨层 Go 授权核心层（license-claims v2 → model/auth）

状态: 完成（可编译验证，go test 全绿）
owner: 协调者（Lead）+ 用户授权安装 Go 工具链
base: 基线 b9a8076（工作树干净，仅新增 4 个 Go 文件）

## 背景
已冻结 license-claims v2（W03）与 features.go 生成物（W02b），但服务端 handler/auth
无授权检查。本轮在用户授权安装 Go 工具链后，实现授权核心层。

## 改动（全部为只增）
- 安装 Go 工具链：下载官方 go1.26.0 darwin/amd64 到 /tmp/go（本机无 brew，MacPorts
  无 golang 主包；官方 tarball 可恢复、安全）
- 新建 `apps/server/internal/model/license.go`
  - Claims/Grant 结构（对齐 license-claims.json）
  - grantActive：active 且窗口左闭右开
  - HasFeature：自身 grant active + 全部依赖满足（visited 防环）
  - MissingDeps：诊断未满足的直接依赖
  - 复用 features.go 生成物的 FeatureDeps（R42.2 单一来源，不重复定义）
- 新建 `apps/server/internal/model/license_test.go`（6 个测试）
- 新建 `apps/server/internal/auth/feature.go`
  - CheckFeature(c *model.Claims, key, now)：纯函数授权判定（默认拒绝）
- 新建 `apps/server/internal/auth/feature_test.go`（4 个测试）

## 证据
- `go vet ./...` → exit=0
- `go test ./...` → auth/config/handler/migration/model 全 ok
- 新增 10 个测试全 PASS（active/依赖缺失/状态拒绝/窗口边界/依赖环防递归/nil 拒绝）
- node 90/90 + python 综合门禁 OK（无回归）
- 提交 5e881e3，工作树干净

## 没有做/限制（如实）
- 未接入 handler 具体端点：需 license 数据源（DB 表或签发方 API 提供 Claims），
  属下一步工作项
- 未做 JWS 签名验证（R42.8：需真实签发方私钥，验证不在产品代码内）
- 未做 DB 持久化 license claims（后续 schema + 加载）

## 下一项
- handler 接入：license claims 数据源（DB license 表 + 加载 → Deps.ClaimsFunc）
  → listContactsV2 等业务端点接入 CheckFeature
- 或先跑服务端既有 DB 集成测试（需 TEST_DATABASE_URL_HANDLER/PG 环境）
