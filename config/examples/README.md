# 配置片段使用约束

两个JSON仅对应当前官方DSH MCP插件的config类型，不是完整宿主配置，不包含插件加载、OAuth交互或真实命令。保留failOnStartupError=true以避免初连失败被误当接入成功；这不取代调用时鉴权。

先按[接入说明](../../docs/install/DSH与Agent-Skills-MCP接入.md)核实安装版本和安全凭证机制，再在受控环境填写真实URL/二进制路径。headers/env不携带秘密；空headers遇受保护服务器应401，而不是通过修改服务端解除认证。

示例中的.invalid域、__VERIFIED占位符均不可用于正式运行。开发包校验仅检查无秘密/形状，不声称网络、OAuth或bridge已可用。生产密钥和用户token不写回本包。
