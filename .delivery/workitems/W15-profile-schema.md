# W15 串行共享窗口：AI 画像契约冻结（packages/profile-schema）

状态: 完成（纯新增 JSON + Python 校验器 + 反证测试）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改文件未触碰（router 为 W13 基线修改）

## 背景
packages/profile-schema 原为空包（packages/README 声明为 "AI 画像 Schema"）。
本轮冻结 R05 范围最后一个空包合同缺口。

## 改动（全部为只增）
- 新建 `packages/profile-schema/profile-schema.json`
  - schema_version/profile_id/tenant_id/subject_ref/source/layers/generated_at/derived
  - layers 枚举：identity/facts/evidence/snapshots（分层，人工新版本不被 AI 覆盖）
  - derived：rebuilt/from/unique_source（必须 false——画像是可重建派生非唯一数据源）
  - notes 对齐文档措辞：画像是可重建派生、受限 Worker 生成、原始正文与 AI 派生分层
- 新建 `scripts/verify_profile_schema.py`
  - 结构校验：必填字段、layers 枚举子集、derived 必填、示例 unique_source=false 强约束
- 新建 `packages/profile-schema/examples/profile-schema.valid.json`
- 更新 `packages/contracts/README.md`（登记 profile-schema + 校验入口）

## 证据
- `python3 scripts/verify_profile_schema.py` → OK + 1 example, exit=0
- 坏例：derived.unique_source=true → FAIL, exit=1；恢复后通过
- 全量回归：6 个 Python 校验器全绿（feature/contract/protocol/source/profile/generator）
  + node 79/79
- 用户 11 处已修改文件未触碰

## 没有做/限制（如实）
- 未在 Go/Rust/TS 中实现 AI 画像生成/消费（跨层真实编码，P0 阻塞）
- 未做签名/时间单位/身份绑定语义校验（R42.8，需真实签发代码）
- 未更新 MANIFEST.sha256（保持既有口径）

## 下一项
- 合同线已全部收口：contracts/ + protocol/ + source-contract/ + profile-schema/
  四类均冻结，packages 空包缺口清零
- P1 跨层真实编码（SourceAdapter/AI 画像/服务端 API/心跳/ingestion）仍等用户决策
