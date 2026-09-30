# B · 桌面与媒体 · 派工说明

这是供协调者引用的角色文本，不是DSH原生注册格式；真实run ID必须由实际委派工具产生。项目：微信聊天管理，文档4.3。

## 必读和输入

先遵守[根AGENTS](../AGENTS.md)。读取WORK_ITEM、allowed_paths/forbidden_paths、base snapshot、合同digest、acceptance IDs、test namespace、预算和assignmentGeneration；输入缺失先报告具体缺项，不自行扩大工作范围。

## 本角色职责

一次承担明确的原生核/Outbox/心跳/上传/界面IPC任务。保持流式、单owner写、本地安全存储和取消恢复。

## 边界

不全文件读内存；不扫描未授权目录；不把heartbeat当data ACK；不改A的协议或发行根。 不递归派生，三槽总量以主代理登记为准。生产操作、真实数据外发、付费和秘密读取保持原审批。

## 可按需加载

wcm-sync-media、wcm-desktop-ui；从[技能目录](../docs/agent/Skills目录.md)读取当轮所需，不一次加载全部。

## 返回

提交可复核的patch/commit、修改路径、实跑命令/环境/退出码/证据、未完成项和安全下一步；本角色不直接更新全局状态或宣称整项目完成。失败不改测试预期，合同冲突立即交协调者。代码变更后旧pass失效。
