# 项目交接记录 — chatlog-web

INSTRUCTION_REVISION: WCM-DELIVERY-V4.3
交接时间: 2026-09-30 22:20
分支: main | HEAD: b9ca5a2 | 工作树干净

## 当前状态（已上线运行）
- **运行中 chatlog-server(8080) 已换新镜像 deploy-server:latest(8e3adb)**，含 W24-W33 全部 license/device/heartbeat 端点
- 正式库(chatlog)迁移已应用 001-004；license claims revision 9 已加载（present:true, archive.read Allowed）
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

## 全量验证证据（2026-09-30 22:15 终检）
- Go 全量真实 PG（chatlog_test）：auth/config/db/handler/migration/model 全 ok
- node：90/90 PASS
- Python：verify_contract/protocol/feature_catalog + a14 全 OK
- admin-web：lint 0 errors（14 存量 warnings）+ vue build DONE
- desktop：cargo test 2/2 PASS
- 正式服务端到端：注册→设备→心跳→license status 全 PASS

## 已知限制/未做（如实）
- 正式库 license claims **未签名**；运行容器未配 `CHATLOG_LICENSE_VERIFY_KEY` → 走直接解析。
  配置密钥即启用 JWS 验签（届时未签名 claims 默认拒绝，属预期强约束）
- 发证/真实签名签发属发行方域（需独立 Issuer 私钥，不进入产品/源码/日志）
- 前端 ContractsView 消费真实 status：需改 api 层（基线已提交，现可做但未做）
- admin-web 14 条 lint warnings（存量，非本次引入）

## 下一步（按优先级）
1. 前端 ContractsView 接入真实 license status（api 层改造）
2. 如需强验签：发行方签发 JWS → 写入 claims_jws → 配置 CHATLOG_LICENSE_VERIFY_KEY
3. 生产发布验收（需干净环境按说明重做安装/升级；真实 PG/TB/发证验收）
4. 三槽并行若恢复：worktree 隔离已解锁（用户 63 处未提交已提交基线）

## 安全/密钥
- 未在本交接写入任何明文密钥/密码（数据库/签发私钥均不在文档）
- 运行容器 DATABASE_URL 含密码（继承原 compose 环境，未改动）
- 回滚：旧镜像 4e9119 可用；迁移 003/004 幂等（IF NOT EXISTS），可安全保留