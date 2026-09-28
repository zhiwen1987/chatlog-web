# ADR-001: monorepo

状态：已接受
日期：2026-09-28

## 背景

规格（§5）要求最终目录为 `chatlog-enterprise/` Monorepo：
`apps/{admin-web,server,desktop-client}` + `packages/*`。
现有 `chatlog-web`（Vue 3 单页应用）必须整体迁入 `apps/admin-web/`，功能不得改变（规格 §2.1）。

## 决策

- 采用 **npm workspaces**（`workspaces: ["apps/*", "packages/*"]`），根 `package.json` 提供聚合 `build/lint/test/serve`。
- 现有 Web 项目整体 `git mv` 至 `apps/admin-web/`，保留 git 历史与 Apache-2.0 License。
- 包名改为 `@chatlog/admin-web`。
- 建立骨架：`apps/server`（Phase 2 Go）、`apps/desktop-client`（Phase 5 Tauri）、`packages/{ui,protocol,shared-types,source-contract,profile-schema}`。
- 不删除任何旧功能、旧测试、旧文档。

## 影响

- 开发时 `npm run build -w @chatlog/admin-web` 或根 `npm run build`。
- 旧路径全部前缀 `apps/admin-web/`；脚本用相对路径定位自身（如 `tests/*.py` 用 `Path(__file__)`），迁移后自动正确。
- 回归验证：`test` 45/45、`lint` 通过、`build` 通过、浏览器冒烟 29+36 项通过。