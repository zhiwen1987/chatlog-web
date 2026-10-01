-- 007_media_objects.sql
-- 媒体对象上传登记（R42.7 upload bytes 级）。
-- 对象 bytes 存对象存储（minio）；此表只登记元数据（object_ref/sha256/size），
-- 供对账与审计，不复制 bytes。上传以 sha256 为内容寻址幂等键：
-- 同 (tenant_id, sha256) 重复上传返回既有 object_ref（不重复写对象）。
-- 对象/元数据分离（AGENTS A06）：bytes 在 minio，元数据在此，丢失可分别重建/对账。

CREATE TABLE IF NOT EXISTS media_objects (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  object_ref  TEXT NOT NULL,
  sha256      TEXT NOT NULL,
  media_type  TEXT NOT NULL,
  size_bytes  BIGINT NOT NULL DEFAULT 0,
  uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, sha256),
  UNIQUE (tenant_id, object_ref)
);

CREATE INDEX IF NOT EXISTS idx_media_objects_tenant_uploaded
  ON media_objects (tenant_id, uploaded_at DESC);