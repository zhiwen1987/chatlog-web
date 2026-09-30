-- 006_ingest_message_idempotency.sql
-- 消息接入幂等（R42 数据链 ingestion）。
-- 同一租户内按 upstream_message_id 唯一去重（NULL 允许多行：PG 唯一索引默认允许多 NULL）。
-- ACK 语义：接收方仅在该行耐久提交后确认；upstream_message_id 缺失的消息不幂等（各为独立行）。

CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_tenant_upstream
  ON messages (tenant_id, upstream_message_id)
  WHERE upstream_message_id IS NOT NULL;