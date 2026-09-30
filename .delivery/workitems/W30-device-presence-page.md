# W30 跨层：admin-web 设备在线页（presence 前端展示）

状态: 完成（lint 0 errors，90 node 测试 PASS，vue build OK）
owner: 协调者（Lead）串行+全自动推进
base: main 9c9c004（工作树干净）

## 改动
- `apps/admin-web/src/api/enterprise.js`：新增 `getDevices()`（GET /api/v1/devices，
  复用服务端已有 listDevices 端点，enterprise 通道带 Bearer token）
- `apps/admin-web/src/views/Devices.vue`：新设备页
  - 表格列：设备名/id 短码、平台、架构、客户端版本、状态、最后心跳
  - 在线判定：last_seen_at 距今 < 5 分钟（ONLINE_WINDOW_MS）→ 在线/离线
  - 完整空态/错误态/加载态；仅 enterprise 模式可用（否则引导去数据来源）
- `apps/admin-web/src/router/index.js`：`/devices` 路由（title 设备）
- `apps/admin-web/src/layout/index.vue`：导航"我的档案"下加"设备"项（icon: server）

## 证据
- `vue-cli-service lint --no-fix`：0 errors（14 个 warning 均为既有 ownership.js 等，
  非本次改动引入）
- `node --test tests/*.test.mjs`：90/90 PASS（基线未动）
- `vue-cli-service build`：DONE Build complete（Devices.vue 已编入 dist）
- 提交 9c9c004，工作树干净

## 没有做/限制
- 未接"设备下线/禁用"操作（仅展示状态，禁用手动操作需审批边界）
- 未做心跳实时推送（页面手动刷新/加载时读取，最小充分）
- 运行中 chatlog-server 容器未重启（新端点/页面接入需重建镜像+重启，属需授权动作）

## 下一项
- 服务端镜像重建：让 license/device/heartbeat 端点在运行中 chatlog-server 上线
  （需用户授权动运行容器，A08 边界）