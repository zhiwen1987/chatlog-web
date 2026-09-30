# W24 串行共享窗口：license status 只读端点（handler 接入）

状态: 完成（跨层 Go，编译+测试全绿）
owner: 协调者（Lead）全自动推进
base: main 5e881e3（W23 auth/model 层）

## 改动
- `apps/server/internal/handler/server.go`
  - Deps 加 `License *model.Claims`（可空；nil=默认拒绝，R42.7）
  - router 加 `GET /api/v1/license/status`
- 新建 `apps/server/internal/handler/license.go`
  - `licenseStatus`：只读展示许可状态摘要（present/mode→licensee+deployment/features 判定）
  - 用 `auth.CheckFeature` 判定关键 feature（archive.read/ingest/media.upload.image/preview）
  - 不验证 JWS 签名（R42.8：需签发方）
- 新建 `apps/server/internal/handler/license_test.go`
  - 3 个测试：nil claims（present=false + 拒绝）、空 claims（拒绝）、
    有 claims（archive.read 允许 + media.upload.image 因缺依赖拒绝）

## 证据
- `go test -count=1 ./internal/handler/ ./internal/model/ ./internal/auth/` → 全 ok
- `go test ./...` 全 ok；`go vet ./...` 无警告
- node 90/90；综合终检 OK（无回归）
- 调试修复：model.Claims 无 Mode 字段（license-claims.json 有 mode，Go 模型未含）→ 移除引用

## 没有做/限制
- 未接 license claims 持久化（DB 表/签发方 API），Deps.License 由调用方提供
- 未做 JWS 签名验证（需真实签发方）
- 未跑 DB 集成测试（需 PG 环境）
