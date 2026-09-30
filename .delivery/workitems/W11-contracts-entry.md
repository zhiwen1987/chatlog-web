# W11 串行共享窗口：契约前端统一入口（admin-web src/lib/contracts.js）

状态: 完成（纯新增 JS + node:test 真实通过）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改 + 63 处未跟踪未触碰（本轮只新增 2 个文件）

## 改动（全部为只增）
- 新建 `apps/admin-web/src/lib/contracts.js`
  - 聚合 W05-W08 四个工具（integrity/license/media/ownership）为统一命名空间
  - `validateContract(type, obj, ...rest)`：按类型分发到对应校验函数
    - integrity-report → countErrors(counts)；license-claims → v2 grants 检查
    - media-manifest → manifestErrors；media-receipt → receiptErrors
    - 未知类型拒绝（不静默通过）
  - 逻辑仍各自封装在源模块，本文件只 re-export + 分发（不复制）
- 新建 `apps/admin-web/tests/contracts.test.mjs`
  - 用 pathToFileURL 加载 contracts.js（data URL 无法解析相对 re-export 导入）

## 证据
- `node --test tests/contracts.test.mjs` → pass 6 / fail 0
- 全量回归 `node --test tests/*.test.mjs` → tests 79 / pass 79 / fail 0
- 调试过程修复：data URL 加载 re-export 模块报 ERR_UNSUPPORTED_RESOLVE_REQUEST，
  改用 pathToFileURL 解决；receiptErrors 在 validateContract 中未定义，改 import 后解决
- 用户 11 处已修改 + 63 处未跟踪未触碰

## 没有做/限制（如实）
- 未在 .vue 视图挂载 validateContract（视图为已修改文件，不触碰；此入口为未来 UI 集成铺路）
- 未做 JWS 签名/时间单位/身份绑定/防回滚/配额语义校验（R42.8，需真实签发代码）
- 未更新 MANIFEST.sha256（保持既有范围差异口径）

## 下一项
- 串行窗口安全工作已到边界：W02-W11 共 11 项收口，79 测试全绿
- P1 跨层真实编码（server presence / Rust heartbeat / ingestion）仍等用户处理
  63 处未提交内容解锁 worktree 并行，或授权主工作树直改