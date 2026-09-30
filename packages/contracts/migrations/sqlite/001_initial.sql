-- 001_initial.sql (SQLite 兼容骨架)
-- 对齐 apps/server/internal/migration/sql/001_initial.sql 的核心表子集。
-- 差异说明：
--  - 无 pgcrypto/pg_trgm/pgrole 扩展；UUID 用 TEXT 存 hex，应用层生成
--  - TIMESTAMPTZ → TEXT(UTC ISO-8601) / 或 INTEGER 毫秒，统一由应用层约定
--  - 无 RLS（SQLite 本地单写 owner）；租户/成员隔离由应用层 WHERE + 复合键保证
--  - 本文件是骨架：仅核心表可执行；其余表（contacts/customers/audit 等）
--    待后续迁移工作项全量归位
-- 验证：sqlite3 :memory: ".read 001_initial.sql" 应无错误

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS tenants (
  id         TEXT PRIMARY KEY,              -- UUID hex（应用层生成）
  name       TEXT NOT NULL,
  status     TEXT NOT NULL DEFAULT 'active'
             CHECK (status IN ('active', 'suspended', 'closed')),
  created_at TEXT NOT NULL,                 -- UTC ISO-8601
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
  id            TEXT PRIMARY KEY,           -- UUID hex
  email         TEXT,
  phone         TEXT,
  name          TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  status        TEXT NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'disabled', 'invited')),
  created_at    TEXT NOT NULL,
  updated_at    TEXT NOT NULL,
  CHECK (email IS NOT NULL OR phone IS NOT NULL),
  UNIQUE (email),
  UNIQUE (phone)
);

CREATE TABLE IF NOT EXISTS tenant_members (
  id         TEXT PRIMARY KEY,              -- UUID hex
  tenant_id  TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role       TEXT NOT NULL DEFAULT 'employee'
             CHECK (role IN ('owner', 'admin', 'supervisor', 'employee', 'analyst', 'auditor', 'device_agent')),
  team_id    TEXT,
  status     TEXT NOT NULL DEFAULT 'active'
             CHECK (status IN ('active', 'disabled')),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (tenant_id, user_id)
);

CREATE TABLE IF NOT EXISTS devices (
  id                 TEXT PRIMARY KEY,      -- UUID hex
  tenant_id          TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id            TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  device_name        TEXT NOT NULL,
  device_fingerprint TEXT NOT NULL,
  platform           TEXT NOT NULL CHECK (platform IN ('windows', 'macos', 'linux')),
  platform_version   TEXT,
  architecture       TEXT NOT NULL CHECK (architecture IN ('x86_64', 'arm64')),
  client_version     TEXT NOT NULL,
  public_key         TEXT NOT NULL,
  status             TEXT NOT NULL DEFAULT 'active'
                     CHECK (status IN ('active', 'disabled', 'revoked')),
  first_seen_at      TEXT NOT NULL,
  last_seen_at       TEXT NOT NULL,
  created_at         TEXT NOT NULL,
  updated_at         TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_devices_tenant ON devices(tenant_id);

CREATE TABLE IF NOT EXISTS source_accounts (
  id           TEXT PRIMARY KEY,            -- UUID hex
  tenant_id    TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  platform     TEXT NOT NULL,
  external_id  TEXT NOT NULL,               -- 来源 external ID（非主键，R42.7 id_kinds）
  display_name TEXT,
  status       TEXT NOT NULL DEFAULT 'active'
               CHECK (status IN ('active', 'disabled', 'revoked')),
  created_at   TEXT NOT NULL,
  updated_at   TEXT NOT NULL,
  UNIQUE (platform, external_id)
);

CREATE TABLE IF NOT EXISTS conversations (
  id           TEXT PRIMARY KEY,            -- UUID hex
  tenant_id    TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  source_id    TEXT NOT NULL REFERENCES source_accounts(id) ON DELETE CASCADE,
  kind         TEXT NOT NULL CHECK (kind IN ('direct', 'group', 'self')),
  external_ref TEXT NOT NULL,               -- 来源会话引用
  title        TEXT,
  created_at   TEXT NOT NULL,
  updated_at   TEXT NOT NULL,
  UNIQUE (source_id, external_ref)
);

CREATE TABLE IF NOT EXISTS messages (
  id               TEXT PRIMARY KEY,        -- UUID hex
  tenant_id        TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  conversation_id  TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  sender_id        TEXT NOT NULL,
  content          TEXT,
  content_type     TEXT NOT NULL DEFAULT 'text'
                   CHECK (content_type IN ('text', 'image', 'video', 'audio', 'file', 'system')),
  seq              TEXT NOT NULL,           -- 单调 seq 十进制字符串（R42.7）
  sent_at          TEXT NOT NULL,           -- UTC ISO-8601
  created_at       TEXT NOT NULL,
  UNIQUE (conversation_id, seq)
);

-- 其余表（contacts/customers/customer_identities/message_observations/
--   audit_log）：待后续迁移工作项按 PG 001 全量归位。
