# W01 串行共享窗口：合同/迁移骨架冻结

状态: 完成（仅骨架，不包含业务实现）
owner: 协调者（Lead）串行窗口
base: HEAD 067c023（未提交改动保留原样）

## 改动
- 新建 packages/contracts/README.md（唯一跨语言合同源声明）
- 新建 migrations/README.md + migrations/{postgres,sqlite,issuer}/（三类库分离）
- 未改动: apps/server/internal/migration/sql/001_initial.sql（保留现状，后续迁移工作项再归位）

## 证据
- find packages/contracts migrations -type f → 3 个文件
- 全部为只增，无覆盖/删除现有文件

## 下一步
- W02 三槽并行（真实编码）：A=server presence/messages 纵切、B=Rust heartbeat client、C=admin-web 设备页状态
- 前提: 各槽独立 worktree/分支 + 测试 namespace 隔离
