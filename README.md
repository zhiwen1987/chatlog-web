# 微信聊天管理

**作者：冰浪｜联系：zhiwen2528@163.com｜新增及二次开发版权：© 2026 进源网络。** 非微信官方产品；上游权利和必要声明保留。

## 这是什么

本包是 **v4.3 项目开发与交付文档包**：围绕跨平台桌面采集、企业服务端、Web管理、图片/视频/语音/文件、TB级档案、客户画像、MCP、安全与按功能批量授权，提供一套可由DSH执行、由工程师接手、由管理员运维的资料。

**当前不是可安装的业务软件。** 本包实际包含治理文档、Skills、角色说明、安装/运维流程、MCP合同与配置片段、交接模板及只读文档校验脚本。不含桌面安装包、服务镜像、产品源代码、生产签名密钥或实际许可证。

## 按你的角色开始

| 角色/目的 | 入口 |
|---|---|
| 我把包交给DSH继续开发 | [00-开始使用](00-开始使用.md) → [AGENTS](AGENTS.md) → [开发手册M01](开发手册.md#m01) |
| 我要了解项目和全部功能 | [项目介绍](项目介绍.md) → [功能说明](功能说明.md) |
| 我要安装软件 | [安装指南](安装指南.md)：先确认拿到真实发布包，再选择客户服务端/桌面/发行方 |
| 我要配置Agent、Skills、MCP | [DSH与接入](docs/install/DSH与Agent-Skills-MCP接入.md) |
| 我接手代码和部署 | [项目交接](项目交接.md) → [维护交接](docs/handoff/维护交接.md) |
| 我负责日常运行与故障 | [运维](docs/operations/运维手册.md) → [排错](docs/operations/日志与排错.md) |
| 我负责验收和签收 | [测试发布门禁](docs/testing/测试与发布门禁.md) → [签收清单](docs/handoff/验收签收清单.md) |

## 三类能力，不混成一个“万能Agent”

- **Agent**决定本轮谁负责、能修改哪里、何时合并；最多3个执行槽，主代理不占第4个编码槽。
- **Skills**提供可重复的工作流程；开发技能在`.agents/skills/`，产品业务技能在`runtime-skills/`。
- **MCP**向已鉴权的Host暴露有限业务工具；它不签发许可证、不执行任意SQL或Shell、不代替审批。

产品行为的唯一规则是[开发规则](开发规则.md)。[开发手册](开发手册.md)规定做事顺序；[AGENTS](AGENTS.md)规定代理边界。其他文档是操作视图，不新增一套冲突算法。

## 本包可以立即验证什么

在解压目录，使用Python 3.10或以上运行：

```bash
python3 scripts/verify_delivery.py
python3 -m unittest discover -s tests -v
```

Windows可将`python3`换为已安装的`py -3`。脚本不联网、不读取聊天、不调用模型、不操作产品数据库；仅检查本包文件。结果不能冒充TB、安装、授权或MCP业务验收。

可选的完整JSON Schema样例检查（仅在受控Python环境安装文档检查依赖后）：

```bash
python3 -m pip install -r requirements-doc-check.txt
python3 scripts/verify_contract_examples.py
```

此依赖只用于文档合同样例，不是桌面客户端或服务端的运行依赖。基础文档校验和41项校验器测试仅使用Python标准库。样例字段校验不证明真实MCP的权限、数据库或传输正确。

完整入口见[文档索引](docs/文档索引.md)，本次实际检查及未测范围见[校验报告](reports/文档校验报告.md)。

## 继续现有项目

现有代码、未提交修改和`.delivery/`必须保留。先合并差异，不直接把本README覆盖已有产品README。来源仓库为`zhiwen1987/chatlog-web`；本次只读核查其README，未运行代码。现有仓库说明的是聊天档案阅读器，不能把未来企业服务端功能当作已有能力。参见[来源记录](docs/provenance/来源与版本.md)。

[变更说明](CHANGELOG.md)｜[二次开发](docs/architecture/二次开发指南.md)｜[安全](SECURITY.md)｜[声明](NOTICE-署名说明.md)
