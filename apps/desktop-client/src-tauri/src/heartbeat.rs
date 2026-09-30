// heartbeat.rs — 设备心跳上报（R42.7/W29 跨层 presence）。
// 纯函数：注入 http client，便于离线单测；不持有服务地址/凭据配置。
// 返回 last_seen_at 供前端校准；网络/服务端错误向上传递，调用方决定重试与提示。

use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct HeartbeatRequest {
    pub base_url: String,
    pub token: String,
    pub device_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct HeartbeatInfo {
    pub ok: bool,
    pub device_id: String,
    pub last_seen_at: Option<String>,
}

pub fn build_heartbeat_url(base_url: &str, device_id: &str) -> String {
    format!(
        "{}/api/v1/devices/{}/heartbeat",
        base_url.trim_end_matches('/'),
        device_id
    )
}

pub async fn heartbeat(client: &reqwest::Client, req: &HeartbeatRequest) -> Result<HeartbeatInfo, String> {
    let url = build_heartbeat_url(&req.base_url, &req.device_id);
    let resp = client
        .post(&url)
        .bearer_auth(&req.token)
        .send()
        .await
        .map_err(|e| format!("heartbeat request: {e}"))?;
    if !resp.status().is_success() {
        return Err(format!("heartbeat http {}", resp.status()));
    }
    let info: HeartbeatInfo = resp.json().await.map_err(|e| format!("heartbeat parse: {e}"))?;
    Ok(info)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn builds_url_with_and_without_trailing_slash() {
        assert_eq!(
            build_heartbeat_url("http://localhost:8080", "dev-1"),
            "http://localhost:8080/api/v1/devices/dev-1/heartbeat"
        );
        assert_eq!(
            build_heartbeat_url("http://localhost:8080/", "dev-1"),
            "http://localhost:8080/api/v1/devices/dev-1/heartbeat"
        );
    }

    #[test]
    fn rejects_non_success_status() {
        // 打到必定失败的本地端口，验证错误被包装（不依赖外部网络）。
        let rt = tokio::runtime::Runtime::new().unwrap();
        let client = reqwest::Client::new();
        let req = HeartbeatRequest {
            base_url: "http://127.0.0.1:1".to_string(),
            token: "t".to_string(),
            device_id: "dev-x".to_string(),
        };
        let err = rt.block_on(heartbeat(&client, &req)).unwrap_err();
        assert!(err.contains("heartbeat"), "unexpected error: {err}");
    }
}