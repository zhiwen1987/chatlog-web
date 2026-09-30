# W31 跨层：Rust 桌面心跳 command（presence 桌面侧）

状态: 完成（cargo check OK，cargo test --lib 2/2 PASS）
owner: 协调者（Lead）串行+全自动推进
base: main e83225e（工作树干净）

## 背景
- 发现本机已有 Rust 工具链（~/.cargo/bin/rustc 1.98.1 + cargo），
  此前"无 rust"记录过时；桌面端 src-tauri 为纯骨架（lib.rs 空、无图标、无 Cargo.lock）

## 改动
- `apps/desktop-client/src-tauri/src/heartbeat.rs`（新增）
  - `build_heartbeat_url` 纯函数（URL 拼接，去尾部斜杠）
  - `async fn heartbeat(client, req)`：POST /api/v1/devices/{id}/heartbeat，
    Bearer token，成功返回 HeartbeatInfo{ok,device_id,last_seen_at}
  - 注入 reqwest::Client 便于离线单测
  - 2 个单测：URL 构造（含/不含尾斜杠）、不可达端口错误分支（真实失败验证错误包装）
- `lib.rs`：`pub mod heartbeat;`
- `Cargo.toml`：声明 `reqwest = 0.13`（default-features=false, json）；
  `dev-dependencies tokio`（单测 runtime）
- `Cargo.lock`（首次提交，桌面可复现构建）
- `icons/`：`tauri icon` 生成完整图标集（修复骨架缺图标致 tauri 宏 panic）
- `.gitignore`：忽略 `/src-tauri/gen/`（Tauri 生成物）

## 证据
- `cargo check`：Finished dev profile（tauri 2.12 + heartbeat 模块编译通过）
- `cargo test --lib builds_url`：1/1 PASS
- `cargo test --lib rejects_non_success`：1/1 PASS
- 提交 e83225e，工作树干净；gen/ 未入库

## 没有做/限制
- 未挂 Tauri command 到窗口前端（桌面端无 UI 逻辑，command 接线待 Phase 5 UI）
- 未跑全量 `cargo test`（Tauri 测试二进制链接极慢，实测 11 分钟未完成；
  心跳单测已用 --lib 隔离跑通）
- 未做设备注册（桌面端需先有注册流程，当前仅心跳上报）

## 下一项
- 服务端镜像重建上线（需授权动运行容器，A08 边界）
- 或 JWS 验签逻辑（自生成测试密钥，逻辑层可落地）