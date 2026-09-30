-- 005_media_manifest_receipt.sql
-- 媒体清单/收据持久化（R42.7 媒体协议对账）。
-- 单一 JSONB 原始对象（manifest/receipt）保持 packages/protocol/*.json 契约为唯一结构源，
-- 不拆多表复制字段；查询列仅索引/对账需要（tenant_id/object_ref/sha256/state）。
-- ACK 语义：receipt 仅代表耐久提交已发生，不代表已发布/索引/媒体已验证（AGENTS A07）。

CREATE TABLE IF NOT EXISTS media_manifests (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id      UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  manifest_id    TEXT NOT NULL,
  object_ref     TEXT NOT NULL,
  sha256         TEXT NOT NULL,
  media_type     TEXT NOT NULL,
  state          TEXT NOT NULL DEFAULT 'stored',
  manifest_json  JSONB NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, manifest_id),
  UNIQUE (tenant_id, object_ref)
);

CREATE INDEX IF NOT EXISTS idx_media_manifests_tenant_state
  ON media_manifests (tenant_id, state, created_at DESC);

CREATE TABLE IF NOT EXISTS media_receipts (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id      UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  receipt_id     TEXT NOT NULL,
  object_ref     TEXT,
  media_kind     TEXT,
  receipt_json   JSONB NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, receipt_id)
);

CREATE INDEX IF NOT EXISTS idx_media_receipts_tenant_created
  ON media_receipts (tenant_id, created_at DESC)
  WHERE object_ref IS NOT NULL;