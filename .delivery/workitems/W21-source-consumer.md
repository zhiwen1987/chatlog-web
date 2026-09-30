# W21 串行共享窗口：SourceAdapter 前端消费工具（admin-web src/lib/source.js）

状态: 完成（纯新增，node:test 全量 87 通过）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改文件未触碰

## 背景
contracts 五个 schema 中 source-adapter 是唯一没有前端消费工具的一个
（integrity/license/media/ownership 均有）。补全合同-前端消费闭环。

## 改动（全部为只增）
- 新建 `apps/admin-web/src/lib/source.js`
  - `sourceErrors(s)`：结构校验（必填字段、capabilities 布尔、cursor.watermark
    十进制字符串、overwrite_policy 非空字符串）
  - `capabilitiesReport(s)`：能力缺失清单（无 cursor 须全量对账）
  - `syncPlan(s)`：游标/覆盖策略展示，is_exhaustive=false 标记 needFullReconcile
    （不冒称可靠 CDC）
- 新建 `apps/admin-web/tests/source.test.mjs`（4 个测试）
- 更新 `apps/admin-web/src/lib/contracts.js`：import+re-export source 三函数，
  validateContract 加 source-adapter 分发
- 更新 `apps/admin-web/tests/contracts.test.mjs`：补 validateContract source-adapter 测试

## 证据
- `node --test tests/source.test.mjs` → 4/4
- `node --test tests/contracts.test.mjs` → 7/7（新增 1 个 source 分发）
- 全量 `tests/*.test.mjs` → 87/87
- 5 个 Python 校验器 + A14 全绿
- 用户 11 处已修改文件未触碰

## 没有做/限制
- 未在 Go/Rust 实现 SourceAdapter 接口（P0 阻塞）
- 未做签名/时间单位/身份绑定语义校验（R42.8，需真实签发代码）
- 未更新 MANIFEST.sha256（保持既有口径）

## 下一项
- 合同-前端消费闭环已补全（5 个 schema 均有消费工具）
- P1 跨层真实编码（SourceAdapter/服务端 API/心跳/ingestion）仍等用户决策
