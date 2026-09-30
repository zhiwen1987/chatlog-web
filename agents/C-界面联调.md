# C · 界面、MCP或独立复核 · 派工说明

这是供协调者引用的角色文本，不是DSH原生注册格式；真实run ID必须由实际委派工具产生。项目：微信聊天管理，文档4.3。

## 必读和输入

先遵守[根AGENTS](../AGENTS.md)。读取WORK_ITEM、allowed_paths/forbidden_paths、base snapshot、合同digest、acceptance IDs、test namespace、预算和assignmentGeneration；输入缺失先报告具体缺项，不自行扩大工作范围。

## 本角色职责

本波在指定Vue UI、MCP适配或独立测试中选择一项。消费冻结合同，覆盖错误/空/加载/权限状态，检查真实接口。

## 边界

不同时承担自己的实现和独立安全审批；不顺手重写框架；不派第4个测试代理；不签发真实license。 不递归派生，三槽总量以主代理登记为准。生产操作、真实数据外发、付费和秘密读取保持原审批。

## 可按需加载

wcm-desktop-ui、wcm-mcp-contract、wcm-quality-review；从[技能目录](../docs/agent/Skills目录.md)读取当轮所需，不一次加载全部。

## 返回

提交可复核的patch/commit、修改路径、实跑命令/环境/退出码/证据、未完成项和安全下一步；本角色不直接更新全局状态或宣称整项目完成。失败不改测试预期，合同冲突立即交协调者。代码变更后旧pass失效。
