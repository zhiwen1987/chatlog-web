# 项目交接记录 — chatlog-web

INSTRUCTION_REVISION: WCM-DELIVERY-V4.3
交接时间: 2026-09-30 22:20
分支: main | HEAD: 21add7e | 工作树干净

## 当前状态（已上线运行）
- **运行中 chatlog-server(8080) 已换新镜像 deploy-server:latest(8e3adb)**，含 W24-W33 全部 license/device/heartbeat 端点
- 正式库(chatlog)迁移已应用 001-006；license claims revision 9 已加载 + claims_jws 已签发（强验签启用）
- 新增 chatlog-admin-web(8081) 承载前端 dist + /api 反代（W37）
- 旧镜像 id 4e9119 留存（回滚手段：`docker run --image 4e9119...` 或 compose 重建）

## 已完成工作链（全部真实执行+证据在 workitems/）
| 阶段 | 范围 | 提交 |
|---|---|---|
| 合同线 W01-W16 | 四类 schema+migrations+校验器+前端消费+UI | dcfa1c1..578cc8e |
| 测试加固 W17-W22 | 90 node 测试+缺陷修复+门禁 | ea5b9d4 |
| Go 授权链 W23-W29 | license model/auth/端点/持久化/设备注册/心跳 | 5e881e3..eeee826 |
| 设备页 W30 | admin-web Devices.vue+api | 9c9c004 |
| 桌面心跳 W31 | Rust heartbeat | e83225e |
| JWS 验签 W32 | SignClaims/VerifyClaimsJWS | 2854700 |
| 验签接入+上线 W33 | 配置驱动验签+镜像重建+容器重启 | 07ddab6,903686f,b9ca5a2 |
| 前端接真实status W34 | ContractsView 接 license/status+4 测试 | 510e573 |
| 契约接后端 W35-W36 | integrity/media 端点+前端接入+迁移005 | d3925a2,dd5751c,563780f,1261e75 |
| 前端上线 W37 | nginx 容器 serve dist+反代 | 7005716 |
| 强验签 W38 | 签发 JWS+配置密钥+重建 server | c083098 |
| 数据链接入 W39 | ingest 端点+授权门禁+幂等+迁移006 | 8cea27a |

## 全量验证证据（2026-10-01 01:40 终检）
- Go 全量真实 PG（chatlog_test）：auth/config/db/handler/migration/model 全 ok
- node：102/102 PASS（W34-W36 +8 测试）
- Python：verify_contract/protocol/feature_catalog + a14 全 OK
- admin-web：lint 0 errors 0 warnings（14 存量已清理）+ vue build DONE
- desktop：cargo test 2/2 PASS
- 正式服务端到端：注册→设备→心跳→license status 全 PASS

## 已知限制/未做（如实）
- ✅ 强验签已启用：claims_jws 已签发（rev9），运行容器已配 `CHATLOG_LICENSE_VERIFY_KEY`+`CHATLOG_LICENSE_AUDIENCE`
  （密钥在 deploy/.env，gitignore 未入库；HS256 对称，RS256 需独立 Issuer 密钥对）
- 生产发布验收阻塞：需真实发行 manifest（TLS/OIDC/证书/域名）+ 发行方审批
- 发证/撤销阻塞：无独立 Issuer 环境（授权中心文档禁止 DSH 代行）
- 三槽并行未恢复：无 3 个可并行独立小项（条件未满足非资源限制）
- 发证/真实签名签发属发行方域（需独立 Issuer 私钥，不进入产品/源码/日志）
- 前端 ContractsView 已接入真实 status（W34，510e573）；演示/未登录态仍展示本地合成 claims 并标注"演示对照"
- admin-web 14 条 lint warnings（存量，非本次引入）

## 下一步（按优先级）
1. ✅ 前端接入真实 license status（W34）、integrity/media（W35-W36）、前端上线（W37）
2. ✅ 强验签启用（W38）
3. ✅ 数据链接入端点（W39）
4. 🔴 生产发布验收（需发行方 manifest + TLS/OIDC + 审批）
5. 🔴 真实发证/撤销（需独立 Issuer 环境）
6. 🔴 upload bytes 级（需 minio 对象存储生产集成）

## 安全/密钥
- 未在本交接写入任何明文密钥/密码（数据库/签发私钥均不在文档）
- 运行容器 DATABASE_URL 含密码（继承原 compose 环境，未改动）
- 回滚：旧镜像 4e9119 可用；迁移 001-006 幂等（IF NOT EXISTS），可安全保留
- 强验签密钥在 deploy/.env（gitignore）；如需轮换：重新签发 claims_jws + 更新 .env + 重建 server