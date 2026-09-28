-- 001_initial.sql
-- Chatlog Enterprise 初始 Schema
-- 覆盖规格 §10-19 核心表 + 设备/分片/审计 + RLS 基础
-- 扩展：uuidv7 用 pgcrypto 生成（gen_random_uuid 为 v4，规格建议消息 id 用 UUIDv7）

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 角色（规格 §63）
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'chatlog_app') THEN
    CREATE ROLE chatlog_app NOLOGIN;
  END IF;
END
$$;

-- ============ 租户与用户 ============

CREATE TABLE IF NOT EXISTS tenants (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  status      TEXT NOT NULL DEFAULT 'active'
              CHECK (status IN ('active', 'suspended', 'closed')),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email         TEXT,
  phone         TEXT,
  name          TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  status        TEXT NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'disabled', 'invited')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT users_email_or_phone CHECK (email IS NOT NULL OR phone IS NOT NULL),
  CONSTRAINT users_email_unique UNIQUE (email),
  CONSTRAINT users_phone_unique UNIQUE (phone)
);

CREATE TABLE IF NOT EXISTS tenant_members (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role       TEXT NOT NULL DEFAULT 'employee'
             CHECK (role IN ('owner', 'admin', 'supervisor', 'employee', 'analyst', 'auditor', 'device_agent')),
  team_id    UUID,
  status     TEXT NOT NULL DEFAULT 'active'
             CHECK (status IN ('active', 'disabled')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, user_id)
);

-- ============ 设备 ============

CREATE TABLE IF NOT EXISTS devices (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id          UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  device_name        TEXT NOT NULL,
  device_fingerprint TEXT NOT NULL,
  platform           TEXT NOT NULL CHECK (platform IN ('windows', 'macos', 'linux')),
  platform_version   TEXT,
  architecture       TEXT NOT NULL CHECK (architecture IN ('x86_64', 'arm64')),
  client_version     TEXT NOT NULL,
  public_key         TEXT NOT NULL,
  status             TEXT NOT NULL DEFAULT 'active'
                     CHECK (status IN ('active', 'disabled', 'revoked')),
  first_seen_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, device_fingerprint)
);

CREATE INDEX IF NOT EXISTS idx_devices_tenant ON devices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_devices_user ON devices(user_id);

-- ============ 数据来源账号 ============

CREATE TABLE IF NOT EXISTS source_accounts (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  source_type         TEXT NOT NULL
                      CHECK (source_type IN ('chatlog_http', 'plaintext_wechat_sqlite', 'manual_file_import', 'windows_wechat', 'macos_wechat', 'linux_wechat', 'enterprise_wechat')),
  external_account_id TEXT NOT NULL,
  display_name        TEXT NOT NULL,
  primary_device_id   UUID REFERENCES devices(id) ON DELETE SET NULL,
  status              TEXT NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active', 'paused', 'disabled')),
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, source_type, external_account_id)
);

CREATE TABLE IF NOT EXISTS source_shards (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  source_account_id   UUID NOT NULL REFERENCES source_accounts(id) ON DELETE CASCADE,
  logical_name        TEXT NOT NULL,
  schema_fingerprint  TEXT NOT NULL,
  source_identity     TEXT NOT NULL,
  status              TEXT NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active', 'archived', 'error')),
  last_seen_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, source_account_id, logical_name)
);

-- ============ 联系人 ============

CREATE TABLE IF NOT EXISTS contacts (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  source_account_id   UUID NOT NULL REFERENCES source_accounts(id) ON DELETE CASCADE,
  external_contact_id TEXT NOT NULL,
  nickname            TEXT,
  remark              TEXT,
  avatar_ref          TEXT,
  contact_type        TEXT DEFAULT 'personal'
                      CHECK (contact_type IN ('personal', 'group', 'official', 'system')),
  metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
  first_seen_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, source_account_id, external_contact_id)
);

-- ============ 客户 ============

CREATE TABLE IF NOT EXISTS customers (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id      UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  display_name   TEXT NOT NULL,
  owner_user_id  UUID REFERENCES users(id) ON DELETE SET NULL,
  team_id        UUID,
  customer_stage TEXT NOT NULL DEFAULT 'lead'
                 CHECK (customer_stage IN ('lead', 'qualified', 'negotiation', 'won', 'lost', 'inactive')),
  status         TEXT NOT NULL DEFAULT 'active'
                 CHECK (status IN ('active', 'archived')),
  manual_note    TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS customer_identities (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id        UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  customer_id      UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
  source_account_id UUID REFERENCES source_accounts(id) ON DELETE SET NULL,
  contact_id       UUID REFERENCES contacts(id) ON DELETE SET NULL,
  identity_type    TEXT NOT NULL,
  identity_value   TEXT NOT NULL,
  verified         BOOLEAN NOT NULL DEFAULT false,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, identity_type, identity_value)
);

-- ============ 会话 ============

CREATE TABLE IF NOT EXISTS conversations (
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id               UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  source_account_id       UUID NOT NULL REFERENCES source_accounts(id) ON DELETE CASCADE,
  external_conversation_id TEXT NOT NULL,
  conversation_type       TEXT NOT NULL DEFAULT 'private'
                          CHECK (conversation_type IN ('private', 'group', 'system')),
  title                   TEXT,
  last_message_at         TIMESTAMPTZ,
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, source_account_id, external_conversation_id)
);

-- ============ 消息 ============

CREATE TABLE IF NOT EXISTS messages (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id          UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  source_account_id  UUID NOT NULL REFERENCES source_accounts(id) ON DELETE CASCADE,
  conversation_id    UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  sender_contact_id  UUID REFERENCES contacts(id) ON DELETE SET NULL,
  direction          TEXT NOT NULL DEFAULT 'unknown'
                     CHECK (direction IN ('incoming', 'outgoing', 'unknown')),
  message_type       TEXT NOT NULL DEFAULT 'text'
                     CHECK (message_type IN ('text', 'image', 'video', 'voice', 'file', 'emoji', 'system', 'link', 'quote', 'unknown')),
  sent_at            TIMESTAMPTZ,
  sent_at_ms         BIGINT,
  content_text       TEXT,
  content_json       JSONB,
  content_hash       TEXT,
  upstream_message_id TEXT,
  decode_status      TEXT NOT NULL DEFAULT 'ok'
                     CHECK (decode_status IN ('ok', 'partial', 'failed', 'encrypted')),
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_messages_tenant_conv ON messages(tenant_id, conversation_id, sent_at);
CREATE INDEX IF NOT EXISTS idx_messages_tenant_time ON messages(tenant_id, sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_content_gin ON messages USING GIN (to_tsvector('simple', coalesce(content_text, '')));

-- ============ 消息观察（幂等/去重核心） ============

CREATE TABLE IF NOT EXISTS message_observations (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id          UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  message_id         UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
  device_id          UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
  source_account_id  UUID NOT NULL REFERENCES source_accounts(id) ON DELETE CASCADE,
  source_shard_id    UUID REFERENCES source_shards(id) ON DELETE CASCADE,
  source_record_key  TEXT NOT NULL,
  source_table       TEXT,
  source_local_id    TEXT,
  upstream_message_id TEXT,
  observed_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  raw_metadata       JSONB NOT NULL DEFAULT '{}'::jsonb,
  UNIQUE (tenant_id, source_record_key)
);

-- ============ 审计日志 ============

CREATE TABLE IF NOT EXISTS audit_log (
  id          BIGSERIAL PRIMARY KEY,
  tenant_id   UUID REFERENCES tenants(id) ON DELETE CASCADE,
  user_id     UUID,
  device_id   UUID,
  action      TEXT NOT NULL,
  resource    TEXT NOT NULL,
  resource_id TEXT,
  detail      JSONB NOT NULL DEFAULT '{}'::jsonb,
  ip_address  TEXT,
  user_agent  TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_tenant_time ON audit_log(tenant_id, created_at DESC);

-- ============ Row Level Security（规格 §65 第二层保护） ============

ALTER TABLE source_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE source_shards ENABLE ROW LEVEL SECURITY;
ALTER TABLE contacts ENABLE ROW LEVEL SECURITY;
ALTER TABLE customers ENABLE ROW LEVEL SECURITY;
ALTER TABLE customer_identities ENABLE ROW LEVEL SECURITY;
ALTER TABLE conversations ENABLE ROW LEVEL SECURITY;
ALTER TABLE messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE message_observations ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;

-- 默认策略：禁止所有直连（应用通过 SET app.tenant_id 后由策略放行）
CREATE POLICY tenant_isolation ON source_accounts USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation ON source_shards USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation ON contacts USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation ON customers USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation ON customer_identities USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation ON conversations USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation ON messages USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation ON message_observations USING (tenant_id::text = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation ON audit_log USING (tenant_id::text = current_setting('app.tenant_id', true));