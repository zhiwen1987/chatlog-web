# W28 跨层：设备注册端点（R42.7，presence/heartbeat 前置）

状态: 完成（真实 PG 集成测试 3/3 PASS）
owner: 协调者（Lead）串行+全自动推进
base: main 77bb46a（工作树干净）

## 改动
- `apps/server/internal/handler/device.go`
  - `POST /api/v1/devices/register`：authed 中间件下，用 JWT 上下文
    （tenant_id/user_id）注册设备
  - 校验：device_name/fingerprint 必填、platform∈{windows,macos,linux}、
    architecture∈{x86_64,arm64}
  - 按 (tenant_id, device_fingerprint) 幂等 upsert，返回稳定 device_id
- `apps/server/internal/db/db.go`：`UpsertDevice`（ON CONFLICT upsert，
  已存在则刷新 last_seen_at 并恢复 active）
- `apps/server/internal/handler/server.go`：路由接线（POST register）
- `apps/server/internal/handler/device_test.go`：3 个集成测试

## 证据
- `go build ./...` exit=0、`go vet ./...` exit=0（/tmp/go/bin/go）
- 真实 PG（chatlog-postgres 测试库）集成测试：
  - TestRegisterDevice PASS（注册成功+幂等返回同 id）
  - TestRegisterDeviceRejectsBadPlatform PASS
  - TestRegisterDeviceRequiresAuth PASS（401）
- 全量 `go test ./...` 真实 PG 全绿（auth/config/db/handler/migration/model ok）
- 提交 77bb46a，工作树干净

## 没有做/限制
- 未做 JWS 签名/密钥绑定校验（需真实签发方）
- 未接 presence/heartbeat 端点（下一步）
- 未跑 node/前端回归（本次纯 Go 改动，node 90 测试基线未动）

## 下一项
- W29 presence/heartbeat 纵切（设备注册已就绪，可建 heartbeat 表+端点+心跳更新 last_seen_at）
