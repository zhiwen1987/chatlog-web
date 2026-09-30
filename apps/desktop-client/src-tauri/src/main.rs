// 桌面客户端 Rust 入口。
// 规格 §3.1 / ADR-002：Tauri 2 + Vue 3 + Rust Native Core + SQLite Local State。
// 骨架阶段仅创建窗口并挂载前端；SourceAdapter、本地状态库、Keychain 等按 Phase 5-7 填充。
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

fn main() {
    chatlog_desktop_lib::run()
}