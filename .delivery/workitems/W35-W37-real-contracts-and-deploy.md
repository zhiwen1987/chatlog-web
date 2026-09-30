# W35-W37 真实契约接入 + 前端 dist 上线

状态: 完成（全量回归全绿 + 部署上线）
owner: 协调者（Lead）全自动推进
base: main a416fc7（工作树干净）

## W35 integrity-report 真实接入
- 服务端 `GET /api/v1/integrity/report`：基于真实 messages 表聚合完整性计数
  （verified/pending/excluded/in_scope/total_discovered/source_missing，租户隔离 WHERE）
- 无新建表（纯聚合查询，YAGNI）；计数满足 R42.10 契约约束
- db.LoadIntegrityReport + handler.integrityReport + 路由
- 集成测试 3/4（Counts 分类/401/空租户零计数）
- 前端 enterprise.getIntegrityReport + ContractsView 完整性区块接真实后端
  （真实数据优先，本地演示标注，错误态显式）
- 提交：d3925a2(server) + dd5751c(frontend)

## W36 media 真实接入
- migration 005：media_manifests + media_receipts 表（JSONB 原始对象 + 查询列，租户隔离幂等）
- 服务端端点：POST /media/manifests、POST /media/receipts、GET /media/receipts（对账列表）
- 校验对齐前端 media.js：必填字段/sha256 64hex/media_type 枚举/object_ref+media_kind 成对
- db.SaveMediaManifest/SaveMediaReceipt/ListMediaReceiptsForReconcile
- 集成测试 4/4（roundtrip 对账/坏 sha 400/不成对 400/401）
- 前端 getMediaReceipts + ContractsView media 区块接真实后端
  （每行 reconcile(receipt, manifest)，无清单/非 stored 拒假 ACK）
- 提交：563780f(server) + 1261e75(frontend)

## W37 前端 dist 部署上线
- deploy/nginx.conf：serve dist + /api 反代 chatlog-server:8080 + SPA fallback + gzip
- deploy/docker-compose.yml：新增 admin-web 服务（nginx:1.27-alpine，8081:80，dist 只读挂载）
- 启动：docker compose up -d --no-deps --no-recreate admin-web
- 端到端验证：静态页 200 / SPA 路由 200 / 注册 201 / license/status 经反代返回真实数据 200
- 提交：7005716

## 全量回归
- Go 真实 PG：auth/config/db/handler/migration/model 全 ok（含 media 4 + integrity 3 + migration 005）
- node：102/102 PASS（W35 +4、W36 +4）
- admin-web：lint 0 errors 0 warnings、vue build DONE
- Python 校验器 + a14：全 OK

## 没有做/限制
- healthz/readyz 未挂 /api 前缀（根路径直连 8080 可用；前端不走它们）
- media 未做对象存储上传（清单/收据为协议层，实际 bytes 存储属 upload 纵切）
- 前端 dist 为构建产物提交范围外（.gitignore 排除，compose 挂载宿主机目录）
