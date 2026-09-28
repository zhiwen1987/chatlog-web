# BASELINE.md — Phase 0 冻结基线

> 建立时间：2026-09-28（本地时区）
> 规格依据：`Chatlog-Enterprise-全平台开发技术规格-v1.0.md` Phase 0
> 原则：只记录、不修改核心代码。以下内容全部为实测结果。
> 更新时间：2026-09-28（用户指示"跑完全程、自行测试修正修复"后，修复了脚本缺陷并重跑全套验证）

## 1. 环境与版本

| 项 | 值 |
| --- | --- |
| 操作系统 | macOS（Darwin，arm64） |
| Node | v24.21.0 |
| npm | 11.19.0 |
| 包管理器 | npm |
| Vue CLI 项目 | `chatlog-web`（@vue/cli-service 5.x，Vue 3） |
| 最近提交 | `577325b docs: reference wcdb-key-tool output compatibility in README introduction` |
| 截图来源提交 | `776fe565ca03e7cd1d41c81fb5cf6ef5c5cc50ae`（`images/*/manifest.json` 记录） |

## 2. 当前功能

### 2.1 数据来源双通道
- **HTTP API 模式**：调用 `/api/v1/{contact,chatroom,session,chatlog}`，兼容 JSON / CSV / 纯文本历史格式；`x-total-count` 分页总数。
- **本地档案模式**：浏览器 OPFS 中真实 SQLite 文件 + SQLite WASM + Zstd 解码，本地导入、查询、导出，不上传后端。
- **演示模式**（demo）：内置虚构数据，离线可演示，不发后端请求。
- 服务地址可在设置中切换（localStorage `chatlog-ui-api-base`，默认生产 `http://127.0.0.1:5030`）。

### 2.2 页面与视图（Vue Router）
8 个桌面页面 + 移动端自适应：
1. `/dashboard` 总览：数据摘要、趋势、最近会话、快捷入口
2. `/chatlog` 聊天记录：会话分栏、搜索、筛选、只读消息、CSV 导出
3. `/analytics` 数据分析：统一时间范围、样本说明、趋势、热力图
4. `/media` 媒体库：图片 / 视频 / 语音 / 文件分类
5. `/contacts` 联系人：目录、搜索、详情弹窗、分页
6. `/chatrooms` 群聊：群列表、成员数、分页
7. `/sessions` 会话：按时间排列、会话过滤
8. `/sources` 数据来源：HTTP / 本地档案 / 演示模式切换

附加视图：`HttpMedia.vue`、`LocalMedia.vue` 媒体详情；`layout/index.vue` 布局。
组件：`MessageContent.vue`、`HttpMessageContent.vue`、`LocalAttachment.vue`、`charts/*`、`ui/*`。

### 2.3 已验证交互能力（来自冒烟测试 manifest）
- 深色主题切换且刷新后保持
- 联系人分页 / 搜索 / 详情弹窗
- 群聊分页、私聊会话过滤
- 后端 demo 搜索与字面量高亮
- CSV 导出真实下载
- 全局键盘搜索走 Vue Router 导航
- 媒体预览打开/关闭
- 移动端 390px 导航、无横向溢出
- demo 模式零后端请求
- 无未捕获浏览器错误
- 真实 API 失败不误激活 demo；空响应显示空状态
- 本地 SQLite：多分片导入、复合 ID 防分片冲突、Zstd 解码、跨页搜索、稳定分页、导出 CSV、附件关联、取消导入保留原档案、非空 WAL 阻止覆盖、清理本地索引

## 3. 测试结果

### 3.1 单元 / 集成测试（Node `node:test`）
测试文件与命令：
- `tests/core.test.mjs`
- `tests/rpc.test.mjs`
- `tests/sql-boundary.test.mjs`
- `tests/wcdb.test.mjs`（npm 脚本 `test:wcdb`）

| 文件 | 结果 |
| --- | --- |
| core.test.mjs | 通过 |
| rpc.test.mjs | 通过 |
| sql-boundary.test.mjs | 通过 |
| wcdb.test.mjs | 通过 |

合计：**45 个用例全部通过**（`npm test`，等价于 `node --test tests/*.test.mjs`）。
**修复**：`package.json` 已新增统一 `test` 脚本 `node --test tests/*.test.mjs`，一次覆盖全部四个测试文件；`test:wcdb` 保留。

### 3.2 浏览器冒烟测试
- 桌面 + 移动端两套：`tests/browser-smoke.py`、`tests/browser-local.py`（Python / Playwright）。
- 分别校验生产构建 `dist/` 下的 UI 与本地 SQLite 档案功能。
- **本次已完整重跑**（用户指示"跑完全程"）：
  - `python tests/browser-smoke.py`：**29 项检查通过，0 浏览器错误**（`images/ui/manifest.json`）。
  - `python tests/browser-local.py`：**36 项检查通过，0 浏览器错误**（`images/wcdb/manifest.json`）。
- 依赖 Pillow 12.3.0（本地已安装）与 Playwright 1.63 Chromium（已装）。
- 冒烟测试会重新生成 `images/ui/` 与 `images/wcdb/` 截图（内容与提交版一致；`sourceCommit` 未设 `GITHUB_SHA` 时 smoke 记 `local`、local 记 HEAD）。本次截图差异为测试再生成的预期产物。

### 3.3 Lint
- **修复**：`package.json` 的 `lint` 脚本已改为 `vue-cli-service lint --no-fix src`（与 README 官方命令一致），规避脚手架枚举 `tests/*.py` 的报错。
- `npm run lint`：**通过（No lint errors found）**。
- 直接 `npx eslint src`：通过（0 错误 0 警告）。

### 3.4 Build
- `npm run build`：通过，产物 `dist/`（约 4.7 MB）。

## 4. 接口

### 4.1 前端调用的 API（`src/api/http.js`）
| 方法 | 路径 | 说明 |
| --- | --- | --- |
| getContacts | `GET /api/v1/contact` | 联系人列表 |
| getChatrooms | `GET /api/v1/chatroom` | 群聊列表 |
| getSessions | `GET /api/v1/session` | 会话列表 |
| getChatLogs | `GET /api/v1/chatlog?format=json` | 聊天记录（分页） |
| getChatLogsRaw | `GET /api/v1/chatlog` | 原始记录 |
| exportChatLogs | `GET /api/v1/chatlog?format=csv` | CSV 导出 |
| getImageUrl | `GET /image/{id}` | 图片 |
| getVideoUrl | `GET /video/{id}` | 视频 |
| getVoiceUrl | `GET /voice/{id}` | 语音 |
| getFileUrl | `GET /file/{id}` | 文件 |
| getDataUrl | `GET /data/{path}` | 本地数据文件 |

请求带 `Accept: application/json, text/csv, text/plain`、`credentials: omit`、12s 超时；响应若为 HTML 会拒绝。分页总数读取 `x-total-count` 头或响应内 `total/count`。

### 4.2 本地档案内部接口（Worker / OPFS）
SQLite WASM 本地导入、分片查询、Zstd 解码、CSV 导出、附件存储，均不经过 HTTP。

## 5. 截图

### 5.1 UI 截图 `images/ui/`
- `dashboard.png` 总览（桌面）
- `chatlog.png` 聊天记录
- `analytics.png` 数据分析
- `media.png` 媒体库
- `media-dark.png` 深色媒体库
- `contacts.png` 联系人
- `chatrooms.png` 群聊
- `sessions.png` 会话
- `mobile-dashboard.png` / `mobile-chatlog.png` / `mobile-media.png`（390px 手机端）
- 说明：全部为虚构演示数据渲染；由 `tests/browser-smoke.py` 对生产 `dist/` 页面生成。

### 5.2 本地档案截图 `images/wcdb/`
- `sources.png` / `sources-dark.png` / `chatlog.png` / `analytics.png` / `media.png` / `mobile-sources.png`
- 说明：实际 SQLite + WASM + OPFS 导入的合成 fixtures 渲染，非模拟。

### 5.3 清单
- `images/ui/manifest.json`：29 项检查，`browserErrors: []`
- `images/wcdb/manifest.json`：36 项检查，`browserErrors: []`

## 6. 基线缺陷（修复情况）

1. ~~**`npm run lint` 失败**~~ **已修复**：`lint` 脚本改为 `vue-cli-service lint --no-fix src`，与 README 官方命令一致，通过。
2. ~~**无统一 `test` 脚本**~~ **已修复**：新增 `test: node --test tests/*.test.mjs`，`npm test` 通过 45/45。
3. **截图再生成会改动工作区**：冒烟脚本在未注入 `GITHUB_SHA` 时把 `sourceCommit` 写成 `local`（本地跑属预期；CI 已通过 `GITHUB_SHA` 规避）。本次两套测试已完整重跑，截图与清单为最新产物。
4. 浏览器冒烟依赖本地 Playwright Chromium 与 Python 环境：**本次已满足并完整运行**（Pillow 12.3.0 已安装）。

## 7. Phase 0 结论

- 原项目完整可用：`npm ci` → `npm test`（45/45）→ `npm run lint`（通过）→ `npm run build`（通过）→ 两套浏览器冒烟（29 + 36 项，0 浏览器错误）。
- 功能、页面、接口、截图已如实记录如上。
- 基线缺陷 1/2 已修复，3/4 已按"跑完全程"要求完成。
- **核心源码未改动**；唯一代码改动是 `package.json` 的脚本定义（`lint`、`test`）。
- 下一步（Phase 1）可在此干净基线上开始，继续复用本基线做回归对照。