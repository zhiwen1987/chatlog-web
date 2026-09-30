# MCP接入、权限与边界

业务真源为规则R14/R15/R24/R33/R42。官方规范当前解析到2026-07-28；协议细节按固定SDK/Schema与实际Host验证，不能根据本页另写假协议。[V06]

## 三层设计

Transport/协议adapter → AuthContext与Policy/Entitlement/Approval → ApplicationService/QueryService/Jobs。MCP只是一种入口，不能绕过Web/REST使用的安全与事务逻辑。

Remote使用Streamable HTTP受TLS与OAuth保护；Local使用stdio和本机已授权身份。业务job和presence/device session不等于MCP协议session。2025-11-25兼容adapter与现代语义分离；只有测过的版本写入支持矩阵。

## 连通不等于有权限

先确定Host、用户/机器主体、数据范围、是否允许内容进入模型，再注册最小scope；token严格验证issuer/audience/用途/期限。不能将客户管理员Cookie、商业license token、DSH开发key混用为MCP Access Token。

每次工具/资源读取仍检查数据ACL；tools/list按scope及功能裁剪，直调隐藏工具也要拒绝。平台授权`mcp.read/write`之外，目标操作的feature/企业开关/依赖/配额/资源和外发政策也要满足。标准历史取回仍有Web/REST保护入口，不能靠购买MCP才可拿回历史。

## 输出与数据最小化

查询默认20条、最高100条；context前后合计不超过200且受256KiB总预算，返回截断、cursor、coverage和asOf。分页token与resource URI同样绑定主体/范围；接口输入不准任选tenant。

原件视频和大文件不base64内联，使用短期受控资源/下载机制；signed URL在有效期内是bearer，停止新URL不等于撤回所有旧URL。强撤销路径必须用真实鉴权访问并压测。

## 高风险操作与长任务

客户合并/拆分、清洗apply、删除、设备禁用与隔离恢复先preview，审批绑定主体、参数/影响hash、版本和TTL。模型传`confirm:true`或MRTR补充输入不等于可信真人审批；需要时跳可信Web批准页。

长任务先耐久保存job再返回handle。双方支持Tasks扩展时映射标准Task；不支持时使用普通业务`jobs_get/jobs_cancel`，不自造标准方法。取消只表示取消请求，不保证已提交副作用回滚。

## 产品不提供的能力

没有任意SQL/Shell/URL/文件读取工具，没有客户MCP批量发证/延期/增额/根密钥轮换，没有生产DB直接覆盖恢复。设备心跳携带命令ID也不允许遥控shell。

聊天、附件、网页、工具说明和日志均有prompt注入风险；即使模型误调用，服务端授权/审批/字节预算仍必须挡住。工具annotation是提示，不是权限机制。[V13]

## 本包提供与未提供

提供[工具设计清单](../../mcp/tool-inventory.json)、[合同示例](../../mcp/合同样例.md)、[联调验收](联调与验收.md)和Host配置片段；不含运行handler、OAuth服务器、实际数据库或真实Token。清单禁止直接作为生产allowlist导入，未完整登记的工具不得发布。
