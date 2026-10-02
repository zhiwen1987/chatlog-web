# W42 桌面端：微信数据导出/备份工具（已授权 · 待数据样本）

状态: 规则已解除（红线放开）｜待数据样本｜技术实施未开工（物理阻塞，如实）
owner: 协调者（Lead）
base: main 478a715（红线解除提交）| 工作树干净 | 与 origin/main 同步

## 背景（已核查事实，2026-10-02）
- 用户要求查看「微信数据库备份逻辑」→ 核查结论：
  - chatlog-web 仓库（含 server/admin-web/desktop-client + 3 worktree）无任何微信数据库读取/备份实现
  - 全桌面代码（.py/.go/.ts/.js/.sh/.md）无 `wechat backup` / `解密微信` / `MSG.db` 相关逻辑
  - desktop-client 仅为 Tauri 2 壳（Vue3 前端 + src-tauri Rust 心跳），不含微信数据读取
- 本机微信数据目录（macOS）：
  - `~/Library/Containers/com.tencent.xinWeChat/Data/Documents/xwechat_files/` 存在
  - 当前仅 `all_users/` 与 `lock.ini`；`all_users/sqlite/` 为空（无 MSG*.db 等聊天库）
  - 推论：本机微信未登录或未同步聊天数据（仅目录结构存在，不可据此判定数据完整）
- chatlog 自身备份方案已存在且完整：`docs/operations/备份恢复与应急.md`（PG 物理备份 + WAL + 对象存储 + 删除抑制账本 + 演练要求），非本工作项范围

## 目标（候选范围，需用户确认其一）
- 方案 A：桌面端读取本机微信加密聊天库并导出（解密 MSG.db → 导出聊天记录/媒体）
  - 技术栈沿用 A03：Tauri 2 + Rust + Vue3；原生核负责采集/文件/Outbox
  - 风险：微信私有库解密涉及密钥材料与逆向协议，属敏感取证领域
- 方案 B：chatlog 服务端自身备份恢复方案 → 已存在文档，**无需新立项**（除非用户指增强/演练）
- 方案 C：其他（用户本地另有微信备份脚本/项目，需用户提供路径或名称）

## 红线（AGENTS A08，2026-10-02 经用户明确授权调整）
- 微信密钥提取：**已由用户（本人设备/本人授权数据/合法用途）明确授权解除禁止**，仍仅限本人设备、本人授权数据、合法使用场景
- 仍禁止：隐蔽监控、任意SQL/Shell/文件读取型MCP、应急万能后门、对其他主体数据的密钥提取
- 实施前仍须：本机出现真实聊天库（MSG.db）样本；技术路径严格保持仅在本人设备/本人数据范围内

## 前置/阻塞（如实，2026-10-02 全容器深度核查）
- 本机微信**从未登录账号**：`xwechat_files/` 下无 `wxid*` 账号目录、无 `db_storage/message`、无 `MSG*.db`
- 仅 `all_users/sqlite/lock.ini`（0B）+ 3 个广告 opt-out 库（`app_data/radium/web/profiles/*/heavy_ad_intervention_opt_out.db`，非聊天数据）
- 方案 A 实施对象（真实聊天库）物理不存在 → 解密/导出**无法真实执行**，不伪造
- 技术可行性仍未验证（微信私有库闭源协议，不做承诺）
- 生产发布/审批边界同 HANDOFF（真实发证/发行方审批）

## 用户侧唯一解锁动作（数据样本）
1. 在本机 macOS 微信**登录账号并保持在线同步**（至少一个会话有聊天记录）
2. 登录后预期出现：`xwechat_files/<wxid>/db_storage/message/` 及 `MSG*.db`
3. 出现后告知 Lead → 立即按 A04 启动技术实施

## 验收（样本就绪后细化）
- 未定：样本出现后补具体行为、数据 owner、测试与证据要求

## 下一步
- ⏳ 等用户本机登录微信产生真实 `MSG*.db` 样本（红线障碍已清除，此数据障碍在用户侧）
- 样本就绪后按 A04 固定循环：读真实源码/样本 → 明确问题与非目标 → 最小实现 → 测试 → 证据