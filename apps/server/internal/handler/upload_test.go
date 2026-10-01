package handler_test

// upload_test.go — 媒体对象上传集成测试（R42.7 upload bytes 级）。
// 覆盖：未配 MediaStore 503；未认证 401；media_type 非法 400；
// 成功上传 200 返回 object_ref/sha256/size_bytes 且 db 登记；
// 同内容重复上传 duplicate=true 且 object_ref 不变（内容寻址幂等）。
// 依赖：真实 PG（TEST_DATABASE_URL_HANDLER）+ 真实 minio（TEST_MINIO_*，缺则跳过）。

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"os"
	"testing"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/handler"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/store"
)

// testMinioStore 连接真实 minio 容器；环境缺失返回 (nil, true) 表示跳过。
func testMinioStore(t *testing.T) (handler.MediaObjectStore, bool) {
	t.Helper()
	endpoint := os.Getenv("TEST_MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	ak := os.Getenv("TEST_MINIO_ACCESS_KEY")
	sk := os.Getenv("TEST_MINIO_SECRET_KEY")
	if ak == "" || sk == "" {
		t.Skip("TEST_MINIO_ACCESS_KEY/SECRET_KEY not set; skipping upload test")
	}
	ms, err := store.NewMinioObjectStore(endpoint, ak, sk, "chatlog-test-media", false)
	if err != nil {
		t.Fatalf("new minio store: %v", err)
	}
	return ms, false
}

// multipartBody 构造 multipart/form-data（media_type + file 字段），返回 body 与 Content-Type。
func multipartBody(mediaType string, payload []byte) (*bytes.Buffer, string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("media_type", mediaType)
	fw, _ := w.CreateFormFile("file", "sample.bin")
	_, _ = fw.Write(payload)
	_ = w.Close()
	return &buf, w.FormDataContentType()
}

// doMultipart 发送 multipart 请求并解析 JSON 响应。
func doMultipart(t *testing.T, url string, body *bytes.Buffer, ct, token string) (int, any) {
	t.Helper()
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", ct)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do upload: %v", err)
	}
	defer resp.Body.Close()
	var out any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode upload: %v", err)
	}
	return resp.StatusCode, out
}

func TestUploadRequiresStore(t *testing.T) {
	ts := setupServer(t) // 无 MediaStore → 503
	token := registerToken(t, ts, "upload.nostore@example.com")
	body, ct := multipartBody("image", []byte("x"))
	code, _ := doMultipart(t, ts.URL+"/api/v1/media/upload", body, ct, token)
	if code != http.StatusServiceUnavailable {
		t.Fatalf("no store: got %d want 503", code)
	}
}

func TestUploadRequiresAuth(t *testing.T) {
	ts := setupServer(t)
	body, ct := multipartBody("image", []byte("x"))
	code, _ := doMultipart(t, ts.URL+"/api/v1/media/upload", body, ct, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("no auth: got %d want 401", code)
	}
}

func TestUploadRejectsBadMediaType(t *testing.T) {
	ms, skip := testMinioStore(t)
	if skip {
		return
	}
	ts := setupServerStore(t, ms)
	token := registerToken(t, ts, "upload.badtype@example.com")
	body, ct := multipartBody("gif", []byte("x"))
	code, _ := doMultipart(t, ts.URL+"/api/v1/media/upload", body, ct, token)
	if code != http.StatusBadRequest {
		t.Fatalf("bad media_type: got %d want 400", code)
	}
}

func TestUploadRoundtripAndIdempotent(t *testing.T) {
	ms, skip := testMinioStore(t)
	if skip {
		return
	}
	ts := setupServerStore(t, ms)
	token := registerToken(t, ts, "upload.ok@example.com")

	payload := []byte("chatlog-media-upload-roundtrip-2026")
	wantSha := sha256.Sum256(payload)
	wantHex := hex.EncodeToString(wantSha[:])

	// 首次上传：200，sha256 匹配，size_bytes 正确，object_ref 前缀 s3://
	body, ct := multipartBody("image", payload)
	code, resp := doMultipart(t, ts.URL+"/api/v1/media/upload", body, ct, token)
	if code != http.StatusOK {
		t.Fatalf("upload: got %d body %v", code, resp)
	}
	m1 := resp.(map[string]any)
	ref1, _ := m1["object_ref"].(string)
	sha1, _ := m1["sha256"].(string)
	dup1, _ := m1["duplicate"].(bool)
	if sha1 != wantHex {
		t.Fatalf("sha256: got %s want %s", sha1, wantHex)
	}
	if m1["size_bytes"] != float64(len(payload)) {
		t.Fatalf("size: got %v want %d", m1["size_bytes"], len(payload))
	}
	if len(ref1) < len("s3://") || ref1[:5] != "s3://" {
		t.Fatalf("object_ref: got %q want s3:// prefix", ref1)
	}
	if dup1 {
		t.Fatalf("first upload should not be duplicate")
	}

	// 同内容重复：duplicate=true 且 object_ref 不变（内容寻址幂等）
	body2, ct2 := multipartBody("image", payload)
	code, resp = doMultipart(t, ts.URL+"/api/v1/media/upload", body2, ct2, token)
	if code != http.StatusOK {
		t.Fatalf("re-upload: got %d body %v", code, resp)
	}
	m2 := resp.(map[string]any)
	ref2, _ := m2["object_ref"].(string)
	dup2, _ := m2["duplicate"].(bool)
	if !dup2 {
		t.Fatalf("re-upload should be duplicate")
	}
	if ref2 != ref1 {
		t.Fatalf("re-upload object_ref changed: %q vs %q", ref2, ref1)
	}
}
