# ADR-003: postgresql

状态：已接受
日期：2026-09-28

## 背景

规格（§4、§9、§41）要求核心数据源必须使用 PostgreSQL（禁止 MongoDB/SQLite/ES 作为唯一库）。
服务端采用 Go + PostgreSQL + S3/MinIO 的 Modular Monolith。
所有核心表必须带 `tenant_id` 并启用 Row Level Security（§65）。

## 决策

- 采用 **PostgreSQL 16** 作为 Source of Truth。
- Migration 用版本化 SQL 文件（`internal/migration/sql/001_initial.sql`），Go embed 内嵌，
  通过 `schema_migrations` 表记录已应用版本，保证幂等。
- 初始 Schema（001_initial.sql）覆盖规格 §10-19：
  tenants / users / tenant_members / devices / source_accounts / source_shards /
  contacts / customers / customer_identities / conversations / messages /
  message_observations / audit_log。
- 启用 `pg_trgm`（Phase 10 搜索基础）与 RLS（tenant_isolation 策略，基于 `app.tenant_id` GUC）。
- Docker Compose 提供 postgres:16-alpine 服务，一键启动。

## 影响

- 所有 API 必须从认证上下文取 tenant_id，禁止前端传租户。
- RLS 作为第二层保护；应用连接需先 `SET app.tenant_id`。
- Migration 变更必须新增版本文件，禁止运行时改表（§80）。
- 集成测试依赖真实 PostgreSQL（Docker），通过 `TEST_DATABASE_URL` 注入。