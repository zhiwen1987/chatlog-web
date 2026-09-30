# W06 串行共享窗口：license-claims 前端消费工具（admin-web src/lib）

状态: 完成（纯新增 JS + node:test 真实通过）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改 + 43 处未跟踪未触碰（本轮未触碰任何既有文件）

## 改动（全部为只增）
- 新建 `apps/admin-web/src/lib/license.js`
  - `grantActive(grant, now)`：grant 状态检查（R42.8）
    仅 state=active 且窗口左闭右开覆盖 now 时有效；非 active 状态/坏窗口拒绝
  - `dependencyCheck(featureKey, grants, catalog, now)`：依赖满足度（R42.2）
    feature 可用当且仅当自身 grant active 且全部 depends_on 可用；返回
    { available, missing }，missing 列出未满足的直接依赖
  - `renderClaims(claims, catalog, now)`：平台展示结构
    - 无 feature_grants（v1 或空）→ trusted=false + note（v2-only 防降级，不冒充激活）
    - 有 feature_grants → trusted=true（结构可信），每个 grant 带
      active/depsSatisfied/missingDeps/usable
- 新建 `apps/admin-web/tests/license.test.mjs`
  - node:test + data:text/javascript 动态加载，与现有测试同风格

## 证据
- `node --test tests/license.test.mjs` → pass 6 / fail 0
- 全量回归 `node --test tests/*.test.mjs` → tests 59 / pass 59 / fail 0
  （新增 license 6 + integrity 8 + 既有 45）
- 调试过程真实修复了一个实现 bug：`renderClaims` 有 feature_grants 时未把
  `trusted` 置 true（初始 false 仅用于 v1/空 claims）；测试暴露后已修复

## 没有做/限制（如实）
- 未接 UI 视图（未在任何 .vue 挂载；Media/Sources 等视图是用户已修改文件，不触碰）
- 未做 JWS 签名验证/时间单位/身份绑定/防回滚/配额语义校验（R42.8，需真实签发代码）
- 未更新 MANIFEST.sha256（保持既有范围差异口径）

## 下一项
- 串行纵切可继续：同模式新增前端消费工具（如 media-manifest/media-receipt 前端校验、
  feature 授权展示整合到现有视图）
- P1 跨层真实编码（server presence / Rust heartbeat / ingestion）仍等用户处理
  未提交内容解锁 worktree 并行