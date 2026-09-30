# migrations

三类库分开编号和执行目标（规则 R05 L276）：
- postgres/ : 客户服务端 PG18 迁移（版本锁）
- sqlite/   : 桌面本地 SQLite 迁移
- issuer/   : 发行方授权中心迁移

现状迁移 SQL（Go embed）位于 apps/server/internal/migration/sql/001_initial.sql，
后续迁移工作项按规则迁入对应目录并编号；当前不复制不删现有 SQL。
owner: 协调者串行窗口；子代理不得分配新全局编号。
