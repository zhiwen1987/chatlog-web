package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
)

// listContactsV2 从 PostgreSQL 返回联系人（Enterprise API，规格 §62）。
// 查询：GET /api/v1/contacts?page=1&page_size=50&q=keyword
func (s *Server) listContactsV2(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ContextTenantID(r.Context())
	if tenantID == "" {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	page, pageSize := pagination(r)
	offset := (page - 1) * pageSize
	q := r.URL.Query().Get("q")

	args := []any{tenantID}
	where := "tenant_id = $1"
	if q != "" {
		args = append(args, "%"+q+"%")
		where += ` AND (nickname ILIKE $2 OR remark ILIKE $2)`
	}
	args = append(args, pageSize, offset)

	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, external_contact_id, nickname, remark, contact_type, first_seen_at, last_seen_at
		FROM contacts
		WHERE `+where+`
		ORDER BY last_seen_at DESC
		LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	type contactRow struct {
		ID               string  `json:"id"`
		ExternalID       string  `json:"external_contact_id"`
		Nickname         *string `json:"nickname"`
		Remark           *string `json:"remark"`
		ContactType      string  `json:"contact_type"`
		FirstSeenAt      string  `json:"first_seen_at"`
		LastSeenAt       string  `json:"last_seen_at"`
	}
	var out []contactRow
	for rows.Next() {
		var c contactRow
		if err := rows.Scan(&c.ID, &c.ExternalID, &c.Nickname, &c.Remark, &c.ContactType, &c.FirstSeenAt, &c.LastSeenAt); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan failed")
			return
		}
		out = append(out, c)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      out,
		"page":      page,
		"page_size": pageSize,
		"total":     len(out),
	})
}

// listConversationsV2 从 PostgreSQL 返回会话（Enterprise API）。
func (s *Server) listConversationsV2(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ContextTenantID(r.Context())
	if tenantID == "" {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	page, pageSize := pagination(r)
	offset := (page - 1) * pageSize

	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, external_conversation_id, conversation_type, title, last_message_at, created_at
		FROM conversations
		WHERE tenant_id = $1
		ORDER BY last_message_at DESC NULLS LAST
		LIMIT $2 OFFSET $3`, tenantID, pageSize, offset)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	type convRow struct {
		ID                string  `json:"id"`
		ExternalID        string  `json:"external_conversation_id"`
		ConversationType  string  `json:"conversation_type"`
		Title             *string `json:"title"`
		LastMessageAt     *string `json:"last_message_at"`
		CreatedAt         string  `json:"created_at"`
	}
	var out []convRow
	for rows.Next() {
		var c convRow
		if err := rows.Scan(&c.ID, &c.ExternalID, &c.ConversationType, &c.Title, &c.LastMessageAt, &c.CreatedAt); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan failed")
			return
		}
		out = append(out, c)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      out,
		"page":      page,
		"page_size": pageSize,
		"total":     len(out),
	})
}

// listMessagesV2 从 PostgreSQL 返回消息（Enterprise API）。
// 查询：GET /api/v1/messages?conversation_id=&page=1&page_size=50&q=keyword
func (s *Server) listMessagesV2(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ContextTenantID(r.Context())
	if tenantID == "" {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	page, pageSize := pagination(r)
	offset := (page - 1) * pageSize
	convID := r.URL.Query().Get("conversation_id")
	q := r.URL.Query().Get("q")

	args := []any{tenantID}
	where := "m.tenant_id = $1"
	if convID != "" {
		args = append(args, convID)
		where += " AND m.conversation_id = $" + strconv.Itoa(len(args))
	}
	if q != "" {
		args = append(args, "%"+q+"%")
		where += " AND m.content_text ILIKE $" + strconv.Itoa(len(args))
	}
	args = append(args, pageSize, offset)

	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT m.id, m.conversation_id, m.direction, m.message_type, m.content_text, m.sent_at, m.sent_at_ms
		FROM messages m
		WHERE `+where+`
		ORDER BY m.sent_at DESC NULLS LAST
		LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	type msgRow struct {
		ID             string  `json:"id"`
		ConversationID string  `json:"conversation_id"`
		Direction      string  `json:"direction"`
		MessageType    string  `json:"message_type"`
		ContentText    *string `json:"content_text"`
		SentAt         *string `json:"sent_at"`
		SentAtMs       *int64  `json:"sent_at_ms"`
	}
	var out []msgRow
	for rows.Next() {
		var m msgRow
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Direction, &m.MessageType, &m.ContentText, &m.SentAt, &m.SentAtMs); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan failed")
			return
		}
		out = append(out, m)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      out,
		"page":      page,
		"page_size": pageSize,
		"total":     len(out),
	})
}

// pagination 解析 page/page_size 参数。
func pagination(r *http.Request) (page, pageSize int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return
}

var _ = sql.ErrNoRows // 占位引用，避免后续裁剪时误删导入