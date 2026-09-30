package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/audit"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
)

// Deps 处理器依赖。
type Deps struct {
	DB          *sql.DB
	JWTSecret   string
	TokenTTLMin int
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