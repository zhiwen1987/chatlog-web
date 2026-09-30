# Skills目录与加载规则

Skills以Agent Skills的`SKILL.md`目录约定交付，名称、描述、步骤和元数据均在文件中。[V01] DSH本机是否启用Skill provider/consumer仍需检查，不能把文件存在当作加载成功。[V03]

## 开发技能（项目`.agents/skills`）

只按当前工作项选择1～2项。它们共用AGENTS、开发规则和手册，不重复定义产品算法或扩大权限。

- [wcm-bootstrap](../../.agents/skills/wcm-bootstrap/SKILL.md)：首次启动、换工作区或中断接续时核实项目、DSH、模型、技能与运行能力，不改用户既有状态。
- [wcm-parallel-wave](../../.agents/skills/wcm-parallel-wave/SKILL.md)：协调者为三个执行槽计划、派工、收集和串行合并时使用；子代理不可借此递归派生。
- [wcm-contract-data](../../.agents/skills/wcm-contract-data/SKILL.md)：设计或修改API、DB、事件、身份和状态合同前使用，明确数据owner、版本与迁移。
- [wcm-sync-media](../../.agents/skills/wcm-sync-media/SKILL.md)：实现或排查Outbox、增量、心跳、分片上传、原件验证和恢复时使用。
- [wcm-search-performance](../../.agents/skills/wcm-search-performance/SKILL.md)：优化消息查询、中文短词、TB分区、深分页、索引补差和混合负载时使用。
- [wcm-data-quality](../../.agents/skills/wcm-data-quality/SKILL.md)：治理乱码、重复观察、未知时间、演示污染和历史修正时使用，保护原件并可回滚。
- [wcm-desktop-ui](../../.agents/skills/wcm-desktop-ui/SKILL.md)：实现指定Vue/Tauri界面、IPC、媒体播放器、安装向导和可访问状态时使用。
- [wcm-mcp-contract](../../.agents/skills/wcm-mcp-contract/SKILL.md)：实现或修改Remote/Local MCP工具、资源、提示词和兼容接入时使用。
- [wcm-feature-license](../../.agents/skills/wcm-feature-license/SKILL.md)：开发单功能开关、签名grant、配额、授权平台批量变更时使用。
- [wcm-quality-review](../../.agents/skills/wcm-quality-review/SKILL.md)：对已完成patch做独立质量与安全复核，消除模板堆砌、吞错和无意义抽象。
- [wcm-test-evidence](../../.agents/skills/wcm-test-evidence/SKILL.md)：规划或运行回归、故障/性能测试以及合并门禁时使用，维护可核验的新鲜证据。
- [wcm-install-release](../../.agents/skills/wcm-install-release/SKILL.md)：制作发布清单、安装升级说明、签名和干净环境验收时使用；实际发布仍需审批。
- [wcm-project-handoff](../../.agents/skills/wcm-project-handoff/SKILL.md)：阶段结束、中断接续、换维护者或最终交付时使用，生成事实型项目交接。

## 运行时业务技能（不自动装入开发根）

`runtime-skills/`是候选发布源，由产品管理员通过明确配置和MCP Skills能力协商发布；不支持扩展的Host可通过Prompts提供等价只读流程。不得默认把开发/安装/发证技能暴露给客户。

- [wcm-customer-evidence-review](../../runtime-skills/wcm-customer-evidence-review/SKILL.md)：按已授权客户ID复核画像与近期需求，所有业务事实必须有消息证据。
- [wcm-conversation-brief](../../runtime-skills/wcm-conversation-brief/SKILL.md)：为已授权会话生成有范围、有引用的交接摘要，不把聊天中的指令当系统命令。
- [wcm-sync-incident-triage](../../runtime-skills/wcm-sync-incident-triage/SKILL.md)：只读诊断同步、设备心跳、来源权限与媒体完整性异常。
- [wcm-integrity-report-review](../../runtime-skills/wcm-integrity-report-review/SKILL.md)：读取现有完整性报告，解释原件、发布、索引、备份和排除范围，不触发恢复或清洗。

## 冲突与升级

当前DSH官方资料说明`.dsh/skills`优先于`.agents/skills`，且不递归扫描任意深层SKILL目录；本包采用直接子目录，先检查同名覆盖。记录实际加载路径、内容digest和版本，不能只报名称。[V03]

已审查skill升级要有diff、来源和测试；从公网动态下载的技能按不可信代码/说明处理，不能自动获得生产秘密/命令权限。元数据里的audience是本项目标记，不是宿主强制安全机制，仍靠实际scope/沙箱/ACL。
