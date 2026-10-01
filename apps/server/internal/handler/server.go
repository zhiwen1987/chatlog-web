package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/audit"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

// Deps 处理器依赖。
type Deps struct {
	DB           *sql.DB
	JWTSecret    string
	TokenTTLMin  int
	License      *model.Claims // 许可 claims v2（可空；nil=默认拒绝，R42.8）
	DeploymentID string

	// LicenseVerifyKey 可选：配置后 licenseStatus 走 JWS 验签实时查库；
	// 未配置保持直接解析（兼容现状）。
	LicenseVerifyKey   []byte
	LicenseExpectedAud string

	// MediaStore 可选：媒体对象存储（upload bytes 级，R42.7）。
	// 未配置时 upload 端点返回 503（对象存储未接入）；配置后流式写对象并登记元数据。
	MediaStore MediaObjectStore
}

// MediaObjectStore 媒体对象存储接口（写路径）。
// 实现：internal/store.MinioObjectStore（minio-go）。
// 接口隔离理由：upload 只依赖 PutObject 语义，便于测试注入 stub；真实实现唯一。
type MediaObjectStore interface {
	// PutObject 流式写入对象，返回 (objectRef 形如 s3://bucket/tenant/name, sha256, sizeBytes, err)。
	PutObject(ctx context.Context, tenantID, mediaType string, r io.Reader) (string, string, int64, error)
}

// Server 组装所有 HTTP 路由。
type Server struct {
	Deps
}

// New 创建 Server。
func New(deps Deps) *Server {
	return &Server{Deps: deps}
}

// Routes 返回根 mux（认证在路由内层接入）。
func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /readyz", s.ready)

	// 公开认证端点
	mux.HandleFunc("POST /api/v1/auth/register", s.register)
	mux.HandleFunc("POST /api/v1/auth/login", s.login)

	// 认证中间件
	authed := auth.Authenticate(s.JWTSecret, http.HandlerFunc(s.router))

	mux.Handle("/api/v1/", authed)
	return mux
}

// router 分发受保护的 API 子路由。
func (s *Server) router(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/api/v1/auth/me":
		s.me(w, r)
	case path == "/api/v1/tenants":
		s.listTenants(w, r)
	case path == "/api/v1/users":
		s.listUsers(w, r)
	case path == "/api/v1/devices":
		s.listDevices(w, r)
	case path == "/api/v1/devices/register" && r.Method == http.MethodPost:
		s.registerDevice(w, r)
	case strings.HasPrefix(path, "/api/v1/devices/") && strings.HasSuffix(path, "/heartbeat") && r.Method == http.MethodPost:
		id := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/devices/"), "/heartbeat")
		s.heartbeat(w, r, id)
	case path == "/api/v1/sources":
		s.listSources(w, r)
	case path == "/api/v1/audit":
		s.listAudit(w, r)
	case path == "/api/v1/contacts":
		s.listContactsV2(w, r)
	case path == "/api/v1/conversations":
		s.listConversationsV2(w, r)
	case path == "/api/v1/messages":
		s.listMessagesV2(w, r)
	case path == "/api/v1/license/status":
		s.licenseStatus(w, r)
	case path == "/api/v1/license" && r.Method == http.MethodPost:
		s.writeLicense(w, r)
	case path == "/api/v1/integrity/report":
		s.integrityReport(w, r)
	case path == "/api/v1/ingest/messages" && r.Method == http.MethodPost:
		s.ingestMessages(w, r)
	case path == "/api/v1/media/manifests" && r.Method == http.MethodPost:
		s.saveMediaManifest(w, r)
	case path == "/api/v1/media/receipts" && r.Method == http.MethodPost:
		s.saveMediaReceipt(w, r)
	case path == "/api/v1/media/receipts":
		s.listMediaReceipts(w, r)
	case path == "/api/v1/media/upload" && r.Method == http.MethodPost:
		s.uploadMedia(w, r)
	default:
		http.NotFound(w, r)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// logAudit 便捷封装。
func (s *Server) logAudit(r *http.Request, action, resource, resourceID string) {
	audit.Log(r.Context(), s.DB, audit.Entry{
		TenantID:   auth.ContextTenantID(r.Context()),
		UserID:     auth.ContextUserID(r.Context()),
		DeviceID:   auth.ContextDeviceID(r.Context()),
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  clientIP(r),
		UserAgent:  r.UserAgent(),
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	return r.RemoteAddr
}
