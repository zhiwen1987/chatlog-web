package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

// registerRequest 注册请求（创建租户+Owner 用户）。
type registerRequest struct {
	Company  string `json:"company"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// register 创建租户和 Owner 用户。
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Company == "" || req.Name == "" || req.Email == "" || len(req.Password) < 8 {
		writeErr(w, http.StatusBadRequest, "company, name, email, password(>=8) required")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "hash failed")
		return
	}

	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "tx begin failed")
		return
	}
	defer tx.Rollback()

	var tenantID string
	if err := tx.QueryRowContext(r.Context(),
		`INSERT INTO tenants(name) VALUES ($1) RETURNING id`, req.Company).Scan(&tenantID); err != nil {
		writeErr(w, http.StatusInternalServerError, "create tenant failed")
		return
	}
	var userID string
	if err := tx.QueryRowContext(r.Context(),
		`INSERT INTO users(email, name, password_hash)
		 VALUES ($1,$2,$3) RETURNING id`,
		req.Email, req.Name, hash).Scan(&userID); err != nil {
		writeErr(w, http.StatusConflict, "email already registered")
		return
	}
	if _, err := tx.ExecContext(r.Context(),
		`INSERT INTO tenant_members(tenant_id, user_id, role) VALUES ($1,$2,'owner')`,
		tenantID, userID); err != nil {
		writeErr(w, http.StatusInternalServerError, "create member failed")
		return
	}
	if err := tx.Commit(); err != nil {
		writeErr(w, http.StatusInternalServerError, "commit failed")
		return
	}

	token, err := auth.SignJWT(s.JWTSecret, userID, tenantID, model.RoleOwner, time.Duration(s.TokenTTLMin)*time.Minute)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sign failed")
		return
	}
	s.logAudit(r, "register", "tenant", tenantID)
	writeJSON(w, http.StatusCreated, map[string]string{"token": token, "tenant_id": tenantID, "user_id": userID})
}

// loginRequest 登录请求。
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// login 登录并签发 JWT。
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	var (
		userID, email, hash string
		tenantID, role      string
	)
	err := s.DB.QueryRowContext(r.Context(), `
		SELECT u.id, u.email, u.password_hash, tm.tenant_id, tm.role
		FROM users u
		JOIN tenant_members tm ON tm.user_id = u.id
		WHERE u.email = $1 AND u.status = 'active'
		ORDER BY tm.created_at ASC
		LIMIT 1`, req.Email).Scan(&userID, &email, &hash, &tenantID, &role)
	if err == sql.ErrNoRows {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	} else if err != nil {
		writeErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	if !auth.VerifyPassword(hash, req.Password) {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, err := auth.SignJWT(s.JWTSecret, userID, tenantID, role, time.Duration(s.TokenTTLMin)*time.Minute)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sign failed")
		return
	}
	s.logAudit(r, "login", "user", userID)
	writeJSON(w, http.StatusOK, map[string]string{"token": token, "tenant_id": tenantID, "user_id": userID, "role": role})
}

// me 返回当前用户信息。
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUserID(r.Context())
	if uid == "" {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"user_id":   uid,
		"tenant_id": auth.ContextTenantID(r.Context()),
		"role":      auth.ContextRole(r.Context()),
	})
}