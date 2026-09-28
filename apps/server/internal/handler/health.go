package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

// health 存活探针。
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "chatlog-server",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// ready 就绪探针：检查数据库连接。
func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.DB.PingContext(ctx); err != nil {
		writeErr(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ready", "db": "ok"})
}

var _ = sql.ErrNoRows // 保留引用，避免未来扩展此处时误删导入