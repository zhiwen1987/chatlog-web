// Desktop client 库入口：创建 Tauri 应用。
// 骨架阶段无自定义 command；Phase 5 起在此挂载 Rust Native Core / SourceAdapter。
#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}