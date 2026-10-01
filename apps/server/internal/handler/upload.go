// upload.go — 媒体对象上传端点（R42.7 upload bytes 级）。
// POST /api/v1/media/upload：multipart/form-data，字段 media_type（枚举）+ 文件体（file）。
// 流程：认证+租户 → 未配 MediaStore 503 → multipart 解析 → 流式写 minio（PutObject，
// 服务端算 sha256）→ db 登记 media_objects（内容寻址幂等，重复上传返回既有 object_ref）
// → 响应 {object_ref, sha256, size_bytes}。
// 未整文件入内存（流式分片；AGENTS A07）。对象失败可重试（幂等键 sha256）。

package handler

import (
	"net/http"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/db"
)

var uploadMediaTypes = map[string]bool{"image": true, "video": true, "audio": true, "file": true}

func (s *Server) uploadMedia(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ContextTenantID(r.Context())
	if tenantID == "" {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	if s.MediaStore == nil {
		writeErr(w, http.StatusServiceUnavailable, "media object storage not configured")
		return
	}
	// 上限 64 MiB（产品附件上限；multipart 流式不整读，但给防护性上限）。
	r.Body = http.MaxBytesReader(w, r.Body, 64<<20)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	mediaType := r.FormValue("media_type")
	if !uploadMediaTypes[mediaType] {
		writeErr(w, http.StatusBadRequest, "media_type 非法（image/video/audio/file）")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "file 字段缺失")
		return
	}
	defer file.Close()

	// 流式写对象存储（服务端算 sha256；不整文件入内存）。
	objectRef, sha256Hex, sizeBytes, err := s.MediaStore.PutObject(r.Context(), tenantID, mediaType, file)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "upload failed")
		return
	}
	// 登记元数据（内容寻址幂等：重复 sha256 返回既有 object_ref，不重复写对象）。
	registeredRef, dup, err := db.SaveMediaObject(r.Context(), s.DB, tenantID, objectRef, sha256Hex, mediaType, sizeBytes)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "register object failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"object_ref": registeredRef,
		"sha256":     sha256Hex,
		"size_bytes": sizeBytes,
		"duplicate":  dup,
	})
}
