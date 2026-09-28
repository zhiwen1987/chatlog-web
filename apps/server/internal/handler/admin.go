package handler

import (
	"database/sql"
	"net/http"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
)

// listTenants 租户列表（Owner/Admin）。
func (s *Server) listTenants(w http.ResponseWriter, r *http.Request) {
	if !hasRole(r, "owner", "admin") {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, name, status, created_at
		FROM tenants
		ORDER BY created_at DESC
		LIMIT 100`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()
	type tenant struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
	}
	var out []tenant
	for rows.Next() {
		var t tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Status, &t.CreatedAt); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan failed")
			return
		}
		out = append(out, t)
	}
	s.logAudit(r, "list", "tenant", "")
	writeJSON(w, http.StatusOK, out)
}

// listUsers 当前租户用户列表。
func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ContextTenantID(r.Context())
	if tenantID == "" {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT u.id, u.email, u.name, u.status, tm.role
		FROM users u
		JOIN tenant_members tm ON tm.user_id = u.id
		WHERE tm.tenant_id = $1
		ORDER BY u.created_at DESC
		LIMIT 200`, tenantID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()
	type userRow struct {
		ID     string `json:"id"`
		Email  string `json:"email"`
		Name   string `json:"name"`
		Status string `json:"status"`
		Role   string `json:"role"`
	}
	var out []userRow
	for rows.Next() {
		var u userRow
		var email sql.NullString
		if err := rows.Scan(&u.ID, &email, &u.Name, &u.Status, &u.Role); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan failed")
			return
		}
		u.Email = email.String
		out = append(out, u)
	}
	writeJSON(w, http.StatusOK, out)
}

// listDevices 当前租户设备列表。
func (s *Server) listDevices(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ContextTenantID(r.Context())
	if tenantID == "" {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, device_name, platform, architecture, client_version, status, last_seen_at
		FROM devices
		WHERE tenant_id = $1
		ORDER BY last_seen_at DESC
		LIMIT 200`, tenantID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()
	type deviceRow struct {
		ID            string `json:"id"`
		DeviceName    string `json:"device_name"`
		Platform      string `json:"platform"`
		Architecture  string `json:"architecture"`
		ClientVersion string `json:"client_version"`
		Status        string `json:"status"`
		LastSeenAt    string `json:"last_seen_at"`
	}
	var out []deviceRow
	for rows.Next() {
		var d deviceRow
		if err := rows.Scan(&d.ID, &d.DeviceName, &d.Platform, &d.Architecture, &d.ClientVersion, &d.Status, &d.LastSeenAt); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan failed")
			return
		}
		out = append(out, d)
	}
	writeJSON(w, http.StatusOK, out)
}

// listSources 当前租户数据来源账号列表。
func (s *Server) listSources(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ContextTenantID(r.Context())
	if tenantID == "" {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, source_type, external_account_id, display_name, status, created_at
		FROM source_accounts
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT 200`, tenantID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()
	type sourceRow struct {
		ID                string `json:"id"`
		SourceType        string `json:"source_type"`
		ExternalAccountID string `json:"external_account_id"`
		DisplayName       string `json:"display_name"`
		Status            string `json:"status"`
		CreatedAt         string `json:"created_at"`
	}
	var out []sourceRow
	for rows.Next() {
		var srow sourceRow
		if err := rows.Scan(&srow.ID, &srow.SourceType, &srow.ExternalAccountID, &srow.DisplayName, &srow.Status, &srow.CreatedAt); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan failed")
			return
		}
		out = append(out, srow)
	}
	writeJSON(w, http.StatusOK, out)
}

// listAudit 当前租户审计日志（Auditor 只读）。
func (s *Server) listAudit(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ContextTenantID(r.Context())
	if tenantID == "" {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT action, resource, resource_id, ip_address, created_at
		FROM audit_log
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT 200`, tenantID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()
	type auditRow struct {
		Action     string `json:"action"`
		Resource   string `json:"resource"`
		ResourceID string `json:"resource_id"`
		IPAddress  string `json:"ip_address"`
		CreatedAt  string `json:"created_at"`
	}
	var out []auditRow
	for rows.Next() {
		var a auditRow
		var rid, ip sql.NullString
		if err := rows.Scan(&a.Action, &a.Resource, &rid, &ip, &a.CreatedAt); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan failed")
			return
		}
		a.ResourceID = rid.String
		a.IPAddress = ip.String
		out = append(out, a)
	}
	writeJSON(w, http.StatusOK, out)
}

func hasRole(r *http.Request, roles ...string) bool {
	role := auth.ContextRole(r.Context())
	for _, want := range roles {
		if role == want {
			return true
		}
	}
	return false
}