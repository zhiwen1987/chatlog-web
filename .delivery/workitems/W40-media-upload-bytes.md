# W40 跨层：媒体对象存储 upload bytes 级（minio 集成）

状态: 完成（Go 全包 PASS + admin-web 102/102 + desktop 2/2 + 端到端对象落盘）
owner: 协调者（Lead）全自动推进
base: main 68f26af | 提交 1238a0e（工作树干净）

## 背景
- HANDOFF 剩余待办第 6 项「upload bytes 级（需 minio 对象存储生产集成）」为最后可自动项
- minio 容器已在 compose 运行（chatlog-minio:9000），凭据在 deploy/.env（MINIO_ROOT_USER/PASSWORD）
- media.go 此前只持久化清单/收据元数据，无对象 bytes 上传

## 改动
- `apps/server/internal/store/minio.go`（新增包）
  - `MinioObjectStore`：minio-go v7 写实现，`PutObject` 流式写对象 + 服务端算 sha256
  - object_ref 形如 `s3://<bucket>/<tenant>/<media_type>/<ts>-<rand>`（与 manifest.bytes_ref 同语义）
  - bucket 幂等 MakeBucket；Content-Type 按 media_type 映射
- `apps/server/internal/handler/server.go`
  - `Deps.MediaStore MediaObjectStore` 接口（写路径隔离，测试可注入 stub）
  - 未配置 store → upload 返回 503（如实，不假成功）
- `apps/server/internal/handler/upload.go`（新增）
  - `POST /api/v1/media/upload`：multipart（media_type + file），64MiB 上限，流式写 minio
  - 服务端算 sha256 → db 登记 → 响应 {object_ref, sha256, size_bytes, duplicate}
- `apps/server/internal/db/db.go`：`SaveMediaObject` 内容寻址幂等（同 sha256 返回既有 object_ref）
- `apps/server/internal/migration/sql/007_media_objects.sql`（新增）
  - `media_objects` 登记表（tenant_id/object_ref/sha256/media_type/size_bytes）
  - 对象/元数据分离（AGENTS A06）：bytes 在 minio，元数据在 PG，丢失可分别重建
- `apps/server/internal/config/config.go` + `cmd/server/main.go`
  - MINIO_ENDPOINT/ACCESS_KEY/SECRET_KEY/BUCKET/USE_SSL 装配；端点配置后启用 upload
- `deploy/docker-compose.yml`：server 段注入 MINIO_* env（默认指向 compose 内 minio:9000）
- `apps/server/internal/handler/upload_test.go`（新增 4 集成测试）

## 测试证据
- `go test ./...`（真实 PG chatlog_test + 真实 minio）全 ok，含 4 个 upload 测试：
  - TestUploadRequiresStore（未配 503）/ TestUploadRequiresAuth（401）
  - TestUploadRejectsBadMediaType（400）/ TestUploadRoundtripAndIdempotent（成功+幂等）
- admin-web：102/102 PASS + lint 0 errors + build DONE（回归未受影响）
- desktop：cargo test --lib 2/2 PASS（回归未受影响）
- Python：verify_protocol_schemas / verify_feature_catalog / a14 全 OK
- 端到端（正式容器，迁移 007 已应用）：
  - 注册 → POST /api/v1/media/upload 真实 28-byte 文件 → 200
  - object_ref=s3://chatlog-media/<tenant>/image/<ts>-<rand>，sha256 匹配
  - minio `/data/chatlog-media/...` 对象真实存在；media_objects 表登记一行

## 已知限制（如实）
- 上传写路径完成；读/删/生命周期/对账由外部流程负责（不在本工作项）
- verify_contract_examples.py 需 requirements-doc-check.txt 隔离环境（环境限制非失败）
- 生产发布验收/发证撤销仍阻塞（真实 manifest/独立 Issuer，见 HANDOFF）

## 接续
- 下一可自动项：无（HANDOFF 三阻塞均为外部材料/审批，如实保留）
- 回滚：迁移 007 幂等；旧镜像 deploy-server 由 compose 重建可回退