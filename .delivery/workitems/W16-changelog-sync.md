# W16 串行共享窗口：CHANGELOG 合同冻结记录（A14 文档同步）

状态: 完成（纯文档新增，全量回归通过）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改文件未触碰

## 背景
CHANGELOG 记录到 4.3（2026-09-29），未记录本轮 W02-W15 的合同/工具冻结。
按 A14（行为/契约变更后更新功能说明），补齐记录。

## 改动（只增）
- 修改 `CHANGELOG.md`：新增 `## 4.3.1 · 2026-09-30 · 串行窗口合同冻结（W02-W15）` 节
  - 合同 schema 清单（contracts/protocol/source-contract/profile-schema/migrations）
  - 校验器清单（5 个 verify + gen --check + a14）
  - 前端消费工具（integrity/license/media/ownership/contracts 统一入口）
  - UI 集成（ContractsView.vue + /contracts 路由）
  - 验证状态（node 79、6 校验器、vue build）
  - 未做项（跨层 Go/Rust、JWS 签名）

## 证据
- CHANGELOG.md 新增节已落盘（python 文件方式更新，未覆盖旧节）
- 全量回归：5 Python 校验器 + A14 exit=0 + node 79/79
- 用户 11 处已修改文件未触碰

## 没有做/限制
- 未更新 MANIFEST.sha256（保持既有口径）
- 未在 docs/install 等派生文档补版本（本次为合同冻结，非安装/升级流程变更）

## 下一项
- 串行窗口合同冻结 A14 文档同步已收口（W02-W16）
- P1 跨层真实编码仍等用户决策（worktree 解锁或授权主工作树直改）
