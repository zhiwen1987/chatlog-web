# W14 串行共享窗口：SourceAdapter 契约冻结（packages/source-contract）

状态: 完成（纯新增 JSON + Python 校验器 + 反证测试）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改文件未触碰；router/index.js 为本轮 W13 基线修改（安全）

## 背景
packages/source-contract 原为空包（BASELINE 声明为 "SourceAdapter 契约"）。
本轮冻结该 R05 范围的最后合同缺口。

## 改动（全部为只增）
- 新建 `packages/source-contract/source-adapter.json`
  - contract_version/source_id/capabilities（detect/schema_probe/stable_id/
    shard_generation/cursor）/identity（kind/external_id）/cursor（watermark/
    is_exhaustive）/unsupported/overwrite_policy（low_water_backfill/
    duplicate_observation/account_switch）
  - notes 对齐文档措辞：来源只读、毒数据隔离不假 ACK、无完整变更序列
    不冒称可靠 CDC
- 新建 `scripts/verify_source_contract.py`
  - 结构校验：必填字段、capabilities/identity/cursor/overwrite_policy 子字段、
    required 一致性、示例 watermark 十进制字符串
- 新建 `packages/source-contract/examples/source-adapter.valid.json`
- 更新 `packages/contracts/README.md`（登记 source-contract + 校验入口）

## 证据
- `python3 scripts/verify_source_contract.py` → OK + 1 example, exit=0
- 坏例：watermark 改 '1.5' → FAIL, exit=1；恢复后通过
- 全量回归：5 个 Python 校验器全绿（feature/contract/protocol/source/generator）
  + A14 exit=0 + node 79/79
- 用户 11 处已修改文件未触碰

## 没有做/限制（如实）
- 未在 Go/Rust/TS 中实现 SourceAdapter 接口（跨层真实编码，P0 阻塞）
- 未做签名/时间单位/身份绑定语义校验（R42.8，需真实签发代码）
- 未更新 MANIFEST.sha256（保持既有口径）

## 下一项
- 合同线已全部收口：contracts/ + protocol/ + source-contract/ 三类均冻结
- P1 跨层真实编码（SourceAdapter 实现/服务端 API/心跳/ingestion）仍等用户决策
