-- 003_device_heartbeat_index.sql
-- 心跳查询/设备在线状态排序索引（W29 presence/heartbeat）。
-- 复用 devices.last_seen_at 作为"最后心跳时间"，不建冗余 heartbeat 表（YAGNI）：
-- devices 表已含 tenant_id, user_id, last_seen_at，register 端点已刷新该列。
-- 注意：devices 表在 001 未启用 RLS，此处同样不启用（租户隔离由应用层 WHERE tenant_id 保证），
-- 避免引入 RLS 策略意外阻挡现有 INSERT/SELECT。

CREATE INDEX IF NOT EXISTS idx_devices_last_seen
  ON devices (tenant_id, last_seen_at DESC);