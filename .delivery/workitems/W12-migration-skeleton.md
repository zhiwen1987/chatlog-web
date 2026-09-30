# W12 串行共享窗口：迁移骨架补齐（migrations/sqlite + postgres/issuer 占位）

状态: 完成（纯新增，sqlite3 语法验证通过）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改 + 63 处未跟踪未触碰

## 背景
W01 记录声称已建 migrations/{postgres,sqlite,issuer}/，但核对时发现目录不存在
（W01 承诺与落盘不符）。本轮恢复 W01 承诺的迁移骨架。

## 改动（全部为只增）
- 新建 `packages/contracts/migrations/README.md`：迁移骨架说明（唯一来源=
  apps/server/internal/migration/sql/001_initial.sql，按库分离）
- 新建 `packages/contracts/migrations/sqlite/001_initial.sql`：
  SQLite 兼容核心表子集（tenants/users/tenant_members/devices/source_accounts/
  conversations/messages）
  - 去 PG 扩展（pgcrypto/pg_trgm/ROLE）、TIMESTAMPTZ→TEXT(UTC)
  - 无 RLS（本地单写 owner，隔离由应用层 WHERE+复合键）
  - UUID 用 TEXT hex；seq 十进制字符串（R42.7）
  - 其余表（contacts/customers/audit 等）注释标记待后续迁移工作项归位
- 新建 `packages/contracts/migrations/postgres/README.md` + `issuer/README.md`：
  占位（待后续工作项归位，不重复定义）

## 证据
- `sqlite3 :memory: ".read packages/contracts/migrations/sqlite/001_initial.sql"`
  → 无错误，SQLITE-OK
- 全量回归：node 79/79、verify_feature_catalog/verify_contract_schemas/
  verify_protocol_schemas 全绿
- 用户 11 处已修改 + 63 处未跟踪未触碰

## 没有做/限制（如实）
- 未做全量 14 表 SQLite 迁移（contacts/customers/audit 等待后续工作项）
- 未动 apps/server/internal/migration/sql/001_initial.sql（用户既有文件）
- 未更新 MANIFEST.sha256（保持既有口径）

## 下一项
- 串行窗口安全工作再次收口：W02-W12 共 12 项
- P1 跨层真实编码仍等用户决策（worktree 解锁或授权主工作树直改）
