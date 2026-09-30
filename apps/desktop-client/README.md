# apps/desktop-client — Chatlog Enterprise 桌面客户端

> Phase 5 开始填充（Tauri 2 + Vue 3 + Rust Native Core + SQLite Local State）。
> 依据规格 §3.1、§31、§32 与 ADR-002。

## 结构

```text
apps/desktop-client
├── package.json        # npm workspace 包 @chatlog/desktop-client（Vite + Vue 3）
├── vite.config.js
├── index.html          # Vite 入口
├── src/                # Vue 3 前端（Tauri 通过 devUrl / frontendDist 接入）
│   ├── main.js
│   └── App.vue
└── src-tauri/          # Rust 侧
    ├── Cargo.toml
    ├── build.rs
    ├── tauri.conf.json
    ├── src/main.rs
    ├── src/lib.rs
    └── icons/          # 需要 tauri icon 生成（当前为占位）
```

## 状态

- 当前为**最小可识别骨架**：前端可 `npm install && npm run dev`（Vite 独立跑）。
- Tauri 桌面壳需 **Rust 工具链**（rustc/cargo）后执行 `npm run tauri dev` 编译运行。
- 本机当前无 Rust 工具链，Rust 侧未实际编译验证；结构遵循 Tauri 2 官方模板。
- 图标为占位，正式构建前需用 `npm run tauri icon <源图>` 生成。
- Phase 5 起按 ADR-002 填充：客户端本地库 client-state.db（sources/checkpoints/outbox/upload_batches/failed_records/media_queue/device_config/health_status，§31）、AES-256-GCM 加密（主密钥存系统 Keychain/DPAPI/Secret Service，§32）、SourceAdapter 接口、自动更新。

## 命令

```bash
# 前端开发（无需 Rust）
npm install
npm run dev

# 桌面应用开发 / 构建（需 Rust 工具链）
npm run tauri dev
npm run tauri build
```