# 文档变更记录

## 4.3.1 · 2026-09-30 · 串行窗口合同冻结（W02-W15）

在 4.3 基础上，串行共享窗口内完成合同与工具冻结，未触碰用户未提交内容：

- 合同 schema（唯一源，禁止在 app 目录重复定义）：
  - `packages/contracts/feature-catalog.json`（23 功能/7 依赖/DAG 无环）
  - `packages/contracts/data-ownership.json`（11 Owner 域 + field_mode 十进制字符串）
  - `packages/contracts/license-claims.json`（claims v2，feature_grants 唯一来源）
  - `packages/contracts/integrity-report.json`（counts 一致 + 三态）
  - `packages/protocol/media-manifest.json` / `media-receipt.json`
  - `packages/source-contract/source-adapter.json`
  - `packages/profile-schema/profile-schema.json`
  - `packages/contracts/migrations/`（sqlite 骨架 + postgres/issuer 占位）
- 校验器（正例 + 坏例反证）：verify_feature_catalog / verify_contract_schemas /
  verify_protocol_schemas / verify_source_contract / verify_profile_schema /
  gen_feature_types --check / a14_contract_doc_check
- 前端消费工具（node:test 28 个）：integrity / license / media / ownership /
  contracts 统一入口
- UI 集成：ContractsView.vue + /contracts 路由（vue-cli build 编译通过）

验证：node 全量 79/79、6 个 Python 校验器 + A14 全绿、vue build 通过。
未做：跨层 Go/Rust 实现（无工具链、主工作树有用户未提交内容）、JWS 签名语义校验。


## 4.3 · 2026-09-29

保留4.2三文件治理与全部产品范围；增加Agent角色派工、项目级开发Skills、独立运行时Skills、MCP接入/配置片段/工具设计清单、项目介绍/功能说明、三类安装流程、运维恢复、二次开发与三类交接资料。

新增规则R43和手册M17，只加强交付可追踪性；不修改既有产品、DB、presence、sync、license claims或MCP协议版本。本包验证只涉及文档与合同样例，产品运行状况unknown/not_tested。

## 4.2及之前

输入摘要保存在`docs/provenance/输入摘要.json`。旧版本仅追溯，不与新文件并行充当指令；本包不复制历史全文，不清空任何真实开发进度。
