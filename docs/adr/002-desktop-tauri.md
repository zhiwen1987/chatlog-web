# ADR-002: desktop-tauri

状态：已接受
日期：2026-09-28

## 背景

规格（§3.1、§118 第四原则）要求桌面客户端采用 **Tauri 2 + Vue 3 + Rust Native Core + SQLite Local State**，
支持 Windows/macOS/Linux 三平台。禁止优先使用 Electron。
操作系统客户端支持与具体微信数据格式支持必须分开设计（SourceAdapter 接口）。

## 决策

- 采用 **Tauri 2** 作为桌面客户端框架（体积小、内存占用低、支持系统 Keychain/自动更新）。
- 前端复用 Vue 3（与 admin-web 同栈，共享 packages/ui）。
- Rust 侧承担：文件系统访问、后台同步任务、系统安全存储（Keychain/DPAPI/Secret Service）。
- 本地状态库 `client-state.db`（SQLite）：sources/checkpoints/outbox/upload_batches/failed_records/media_queue/device_config/health_status（规格 §31）。
- 本地敏感数据 AES-256-GCM 加密；主密钥存系统安全存储，禁止明文（§32、§118 第十一原则）。
- SourceAdapter 接口：detect/validate/getAccount/listShards/scanContacts/scanSessions/scanMessages/scanMedia/getCheckpoint/incrementalScan/healthCheck（§7）。

## 影响

- apps/desktop-client 结构：src-tauri（Rust）+ src（Vue）+ client-state.db（本地 SQLite）。
- 平台差异封装在 Rust 侧，Vue 侧只做 UI。
- 自动更新（updater）需配置签名密钥，Phase 5 详细实现。