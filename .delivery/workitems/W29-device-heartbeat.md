# W29 跨层：设备心跳端点（presence/heartbeat 纵切）

状态: 完成（真实 PG 集成测试 4/4 PASS）
owner: 协调者（Lead）串行+全自动推进
base: main eeee826（工作树干净）

## 改动
- `apps/server/internal/migration/sql/003_device_heartbeat_index.sql`
  - 复用 `devices.last_seen_at` 作为"最后心跳时间"，**不建冗余 heartbeat 表**（YAGNI）
  - 新增 `idx_devices_last_seen (tenant_id, last_seen_at DESC)` 供在线状态排序查询
  - 注意：devices 在 001 未启用 RLS，003 同样不启用（租户隔离由应用层 WHERE tenant_id 保证），
    避免引入 RLS 策略意外阻挡现有 INSERT/SELECT
- `apps/server/internal/db/db.go`：`TouchDevice`（租户内按 id 刷新 last_seen_at，
  影响行=0 表示未知/异租户 → 404 语义）
- `apps/server/internal/handler/device.go`：`POST /api/v1/devices/{id}/heartbeat`
  - 归属校验：URL device_id + JWT tenant_id（用户 JWT 不携带设备 did）
  - 返回更新后 last_seen_at；不推进消息 ACK（心跳≠数据收据，AGENTS A07）
- `apps/server/internal/handler/server.go`：路由接线（手工解析 {id} 传参）
- `apps/server/internal/handler/device_test.go`：新增 3 个心跳集成测试

## 证据
- `go build ./...` exit=0、`go vet ./...` exit=0（/tmp/go/bin/go）
- 真实 PG（chatlog-postgres 测试库 chatlog_test）集成测试：
  - TestHeartbeatDevice PASS（注册→心跳→last_seen_at 返回）
  - TestHeartbeatUnknownDevice PASS（404）
  - TestHeartbeatRequiresAuth PASS（401）
  - TestRegisterDevice 系列 PASS（无回归）
- 全量 `go test ./...` 真实 PG 全绿（auth/config/db/handler/migration/model ok）
- 提交 eeee826，工作树干净

## 没有做/限制
- 无独立 heartbeat 表/历史心跳明细（当前仅记录最后心跳时间，符合最小充分）
- 未做"离线判定/超时过期"状态（需业务阈值，非本工作项）
- 未跑 node/前端回归（本次纯 Go 改动，node 90 测试基线未动）

## 下一项
- admin-web 设备在线状态页：接入 api 层展示 last_seen_at / 在线/离线