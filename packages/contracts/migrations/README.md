# 合同迁移骨架

唯一来源: apps/server/internal/migration/sql/001_initial.sql（PostgreSQL 完整版）
本目录为合同层的迁移骨架归档，按库类型分离，禁止在 app 目录重复定义。

## 目录
- `sqlite/001_initial.sql`: SQLite 兼容核心表子集（已可执行，sqlite3 验证通过）
- `postgres/`: PostgreSQL 迁移（后续工作项按 001_initial.sql 归位）
- `issuer/`: 发行方授权中心迁移（后续工作项，独立 PG 身份域）

## 差异说明
- SQLite 版去 PG 扩展（pgcrypto/pg_trgm/ROLE）、TIMESTAMPTZ→TEXT(UTC)、
  无 RLS（本地单写 owner，隔离由应用层 WHERE+复合键）
- UUID 用 TEXT hex，应用层生成（R42.7 十进制字符串仅限 ID/bytes/seq/revision
  数值域，UUID 用 hex 文本）

## 校验
- `sqlite3 :memory: ".read packages/contracts/migrations/sqlite/001_initial.sql"`
