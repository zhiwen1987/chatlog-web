# DSH、Agent、Skills与MCP接入

本页给实际安装核验步骤，不把本项目策略JSON当DSH原生配置。官方资料核验日期2026-09-29；用户安装版、第三方模型路由和插件组合仍需本机测试。[V02,V03,V04,V05]

## 1. 先进入正确工作区

保留当前可用DSH与Flash路由，在Web UI选择真实项目workspace；确认cwd、分支、dirty与旧会话。明确读根AGENTS，并报告`WCM-DELIVERY-V4.3`。官方允许模型设置和工作区选择，不代表本机已自动启用全部委派/MCP能力。

三执行槽按手册M08验证：独立run、路径/合同owner、数据库/对象前缀/端口、取消、结果收集和预算。没有原生能力时不要编造`dsh --parallel 3`或修改不认识的私有参数。

## 2. 安装项目开发Skills

将本包`.agents/skills`的目录合并到项目同名路径；不是复制进聊天，也不是覆盖用户全局Skills。每个直接子目录包含SKILL.md。官方DSH本地provider优先检查`.dsh/skills`，然后`.agents/skills`；存在同名时确认实际来源。[V03]

通过本机提供的Skill目录/加载工具发现`wcm-bootstrap`并打开，记录路径、版本和digest。未出现时检查workspace根、provider/consumer是否加载、同名覆盖和解析错误，不靠写“已开启”解决。不可用时显式读取文件作为受控指令，记录native_skill_loading未验证。

技能使用`name/description`与项目metadata，格式参考Agent Skills；不使用`allowed-tools`假定跨宿主预授权。[V01] 本包runtime-skills为单独产品侧候选，不默认装成开发技能。

## 3. 开发模型与产品模型分开

保留你当前`deepseek-v4-flash`实际可用路由。官方别名服务信息不证明第三方代理真实权重；记录requested/returned model、provider和可见限制。思考/工具参数只按当前provider合同使用，不假设某个temperature或effort设置保证正确。[V11]

产品AI画像、OCR/ASR使用客户自己批准的提供商和凭证，不能借用开发DSH密钥或默认外发真实聊天。

## 4. Remote MCP接入

前置条件：实际服务端已经实现`/mcp`、TLS、OAuth资源保护、scope/ACL、商业入口功能和Host外发政策；不能仅有本包文档就连接。

本包提供[DSH远程config片段](../../config/examples/dsh-mcp-remote.config.fragment.json)。字段来自官方`@deepseek-ai/dsh-mcp-client`的config类型；它只是单插件config对象，**不是完整cordis.yml，也不会自动加载插件**。由管理员在当前DSH实际设置/API/配置机制中挂载，经审查后填写真实URL。[V05]

`headers:{}`故意不放token；未完成OAuth时应401。核验当前Host是否能完成OAuth发现/登录/刷新和正确audience；资料里的通用headers字段不能证明它已实现完整OAuth。缺能力就标blocked并选用经验证的标准接入组件/Host，不能降级成永久共享管理员Key或放宽服务端鉴权。

先用只读测试身份调用system_capabilities/system_me，再验证受限messages_context；没有任何真实聊天外发授权时仅使用合成租户。通过后才增加必要scope；不一次开admin/backup restore。详细见[MCP接入](../mcp/MCP接入与权限.md)。

## 5. Local stdio接入

安装真实已构建的本地bridge并确认绝对路径、签名、版本和参数后，参考[stdio片段](../../config/examples/dsh-mcp-stdio.config.fragment.json)。本包命令占位符不是已存在二进制。默认args/env为空；不可凭模板猜bridge旗标或将长效token写进去。

bridge经受控IPC访问本OS用户授权的原生核；不直接打开浏览器OPFS，不默认监听0.0.0.0，不运行任意shell。stdout只能协议数据，日志stderr且去敏。关闭本地授权后下次调用必须拒绝。

## 6. 运行时Skill与MCP Prompts

支持Skills扩展的Host，产品服务端可按注册allowlist发布`runtime-skills`里的只读业务流程；双方未协商支持则用现有Prompts/普通工具组合，不发送未支持扩展。[V12] 切勿遍历项目磁盘把开发Skill、安装脚本或秘密打包成MCP资源。

## 7. 接入记录

填Host/DSH实际版本、插件版本、代码/规则摘要、transport、OAuth方式、scope/profile、协议版与可选扩展、工具数量和实跑结果。反例包含401、错误audience、跨tenant、写操作无权限、Token刷新失败、断线任务恢复。完整结果未过前标未验收，不称“所有AI通用”。
