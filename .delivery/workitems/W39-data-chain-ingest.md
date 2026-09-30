# W39 数据链纵切：ingestion 端点（授权门禁+幂等）

状态: 完成（4 集成测试 PASS，Go 全量真实 PG 全绿）
owner: 协调者（Lead）全自动推进
base: main 022c5bb（工作树干净）

## 改动
- migration 006：`idx_messages_tenant_upstream` 唯一索引（tenant_id+upstream_message_id，
  WHERE upstream_message_id IS NOT NULL）→ 消息接入幂等
- model.IngestMessage：数据链接入消息（对齐 messages 表 + 来源/会话引用）
- db.IngestMessage：source_account upsert → conversation upsert → message upsert
  （ON CONFLICT DO NOTHING 幂等，返回 accepted/duplicated）
- handler.ingest.go：`POST /api/v1/ingest/messages`
  - 授权门禁：archive.ingest feature（CheckFeature，未授权默认拒绝 R42.8）
  - 枚举校验（source_type/direction/message_type/decode_status 对齐表 CHECK）
  - ACK 语义：accepted>0 表示耐久提交后确认，不推进媒体 ACK
- 集成测试 4/4：无许可 403 / 接入 accepted=1 幂等重发 duplicated=1 / 401 / 枚举 400

## 证据
- 真实 PG（chatlog_test）集成测试 4/4 PASS
- Go 全量真实 PG 全绿（含 migration 006）
- 提交 8cea27a，工作树干净

## 数据链纵切全景
- ✅ ingestion：消息接入（本工作项 W39）
- ✅ oracle 校验：integrity report（W35）+ media 对账（W36）提供数据完整性校验
- ✅ media 元数据：manifest/receipt 端点（W36）
- 🔄 upload bytes 级：真实对象上传到 minio 需对象存储生产集成（后续工作项）

## 没有做/限制
- 未做对象存储上传（media bytes 级，需 minio SDK 分片+校验，独立工作项）
- 未做发布/索引流水线（oracle 的发布环，需检索/索引集成）
- 前端无 ingestion UI（数据入口，非展示）

## 下一步
- upload bytes 级（minio 对象存储集成）
- 三槽并行恢复（worktree 隔离已解锁）
