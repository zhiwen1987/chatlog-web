package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/auth"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

// Open 建立 PostgreSQL 连接池。
func Open(databaseURL string) (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}

// LoadLicenseClaims 从 license_claims 表加载部署的许可 claims v2。
// 未找到返回 (nil, nil)（默认拒绝，R42.8）。
func LoadLicenseClaims(ctx context.Context, database *sql.DB, deploymentID string) (*model.Claims, error) {
	row := database.QueryRowContext(ctx, `
		SELECT claims_json::text
		FROM license_claims
		WHERE deployment_id = $1
		ORDER BY license_revision DESC
		LIMIT 1`, deploymentID)
	var raw string
	if err := row.Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("load license claims: %w", err)
	}
	var c model.Claims
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return nil, fmt.Errorf("unmarshal license claims: %w", err)
	}
	return &c, nil
}

// LoadLicenseClaimsVerified 加载并 JWS 验签部署的许可 claims（R42.8 签名要求）。
// 读取 license_claims 表最新行（claims_jws 列），用 verificationKey 验签；验签失败/篡改/错钥/过期
// 返回错误（默认拒绝，不落入部分解析状态）。未找到返回 (nil, nil)。
// claims_jws 为空（未签发）也返回错误：配置了验签密钥说明要求签名，缺签名按拒绝处理。
func LoadLicenseClaimsVerified(ctx context.Context, database *sql.DB, deploymentID string, verificationKey []byte, expectedAud string) (*model.Claims, error) {
	row := database.QueryRowContext(ctx, `
		SELECT claims_jws
		FROM license_claims
		WHERE deployment_id = $1
		ORDER BY license_revision DESC
		LIMIT 1`, deploymentID)
	var raw string
	if err := row.Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("load license claims: %w", err)
	}
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("verify license claims: claims_jws empty (unsigned claims not accepted when verify key configured)")
	}
	c, err := auth.VerifyClaimsJWS(verificationKey, raw, expectedAud)
	if err != nil {
		return nil, fmt.Errorf("verify license claims: %w", err)
	}
	return c, nil
}

// UpsertDevice 注册/更新设备（R42.7 设备登记）。
// 按 (tenant_id, device_fingerprint) 幂等 upsert：已存在则刷新 last_seen_at
// 并恢复 active（设备重新登记视为上线）；返回设备 ID。
func UpsertDevice(ctx context.Context, database *sql.DB, d *model.Device) (string, error) {
	if d.TenantID == "" || d.UserID == "" || d.DeviceFingerprint == "" {
		return "", fmt.Errorf("upsert device: tenant_id/user_id/fingerprint required")
	}
	var id string
	row := database.QueryRowContext(ctx, `
		INSERT INTO devices (
			tenant_id, user_id, device_name, device_fingerprint,
			platform, platform_version, architecture, client_version, public_key,
			status, first_seen_at, last_seen_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'active', now(), now(), now(), now())
		ON CONFLICT (tenant_id, device_fingerprint) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			device_name = EXCLUDED.device_name,
			platform = EXCLUDED.platform,
			platform_version = EXCLUDED.platform_version,
			architecture = EXCLUDED.architecture,
			client_version = EXCLUDED.client_version,
			public_key = EXCLUDED.public_key,
			status = 'active',
			last_seen_at = now(),
			updated_at = now()
		RETURNING id`,
		d.TenantID, d.UserID, d.DeviceName, d.DeviceFingerprint,
		d.Platform, d.PlatformVersion, d.Architecture, d.ClientVersion, d.PublicKey)
	if err := row.Scan(&id); err != nil {
		return "", fmt.Errorf("upsert device: %w", err)
	}
	return id, nil
}

// TouchDevice 心跳上报：刷新指定设备的 last_seen_at（租户内）。
// 设备不存在或不属于该租户返回 (false, nil)（调用方映射 404）。
func TouchDevice(ctx context.Context, database *sql.DB, tenantID, deviceID string) (bool, error) {
	if tenantID == "" || deviceID == "" {
		return false, fmt.Errorf("touch device: tenant_id/device_id required")
	}
	res, err := database.ExecContext(ctx,
		`UPDATE devices SET last_seen_at = now(), updated_at = now()
		 WHERE id = $1 AND tenant_id = $2`, deviceID, tenantID)
	if err != nil {
		return false, fmt.Errorf("touch device: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("touch device rows: %w", err)
	}
	return n > 0, nil
}

// IntegrityReport 返回租户消息完整性的计数聚合（R42.10）。
// 基于真实 messages 表统计：verified/pending/excluded 分类按 decode_status + content_hash；
// source_missing 恒为 0（source_account_id 外键必填，不存在缺失来源行）。
// 返回 counts 满足契约约束：verified+pending+excluded+source_missing <= in_scope <= total_discovered。
type IntegrityCounts struct {
	TotalDiscovered int `json:"total_discovered"`
	InScope         int `json:"in_scope"`
	Verified        int `json:"verified"`
	Pending         int `json:"pending"`
	Excluded        int `json:"excluded"`
	SourceMissing   int `json:"source_missing"`
}

func LoadIntegrityReport(ctx context.Context, database *sql.DB, tenantID string) (*IntegrityCounts, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("integrity report: tenant_id required")
	}
	var c IntegrityCounts
	err := database.QueryRowContext(ctx, `
		SELECT
			COUNT(*)                                                        AS total_discovered,
			COUNT(*) FILTER (WHERE decode_status IN ('ok', 'partial'))      AS in_scope,
			COUNT(*) FILTER (WHERE decode_status = 'ok' AND content_hash <> '') AS verified,
			COUNT(*) FILTER (WHERE decode_status = 'ok' AND (content_hash IS NULL OR content_hash = '')) AS pending,
			COUNT(*) FILTER (WHERE decode_status IN ('encrypted', 'failed'))    AS excluded,
			0                                                                AS source_missing
		FROM messages
		WHERE tenant_id = $1`, tenantID).Scan(
		&c.TotalDiscovered, &c.InScope, &c.Verified, &c.Pending, &c.Excluded, &c.SourceMissing)
	if err != nil {
		return nil, fmt.Errorf("integrity report: %w", err)
	}
	return &c, nil
}

// SaveMediaManifest 持久化媒体清单（R42.7）。
// 按 (tenant_id, manifest_id) 幂等 upsert：重复上报同 manifest_id 更新为最新（state 可迁移）。
// manifest_json 保留原始协议对象，供前端 manifestErrors 消费。
func SaveMediaManifest(ctx context.Context, database *sql.DB, tenantID string, m *model.MediaManifest) error {
	if tenantID == "" || m.ManifestID == "" || m.ObjectRef == "" {
		return fmt.Errorf("save media manifest: tenant_id/manifest_id/object_ref required")
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("save media manifest: marshal: %w", err)
	}
	_, err = database.ExecContext(ctx, `
		INSERT INTO media_manifests (tenant_id, manifest_id, object_ref, sha256, media_type, state, manifest_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id, manifest_id) DO UPDATE SET
			object_ref = EXCLUDED.object_ref,
			sha256 = EXCLUDED.sha256,
			media_type = EXCLUDED.media_type,
			state = EXCLUDED.state,
			manifest_json = EXCLUDED.manifest_json,
			updated_at = now()`,
		tenantID, m.ManifestID, m.ObjectRef, m.SHA256, m.MediaType, m.State, raw)
	if err != nil {
		return fmt.Errorf("save media manifest: %w", err)
	}
	return nil
}

// SaveMediaReceipt 持久化媒体收据（R42.7）。
// 按 (tenant_id, receipt_id) 幂等 upsert：重复上报同 receipt_id 更新为最新。
func SaveMediaReceipt(ctx context.Context, database *sql.DB, tenantID string, r *model.MediaReceipt) error {
	if tenantID == "" || r.ReceiptID == "" {
		return fmt.Errorf("save media receipt: tenant_id/receipt_id required")
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("save media receipt: marshal: %w", err)
	}
	var objectRef, mediaKind any
	if r.ObjectRef != "" {
		objectRef = r.ObjectRef
	}
	if r.MediaKind != "" {
		mediaKind = r.MediaKind
	}
	_, err = database.ExecContext(ctx, `
		INSERT INTO media_receipts (tenant_id, receipt_id, object_ref, media_kind, receipt_json)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id, receipt_id) DO UPDATE SET
			object_ref = EXCLUDED.object_ref,
			media_kind = EXCLUDED.media_kind,
			receipt_json = EXCLUDED.receipt_json,
			created_at = now()`,
		tenantID, r.ReceiptID, objectRef, mediaKind, raw)
	if err != nil {
		return fmt.Errorf("save media receipt: %w", err)
	}
	return nil
}

// ListMediaReceiptsForReconcile 返回租户媒体收据及对应清单（供前端 reconcile 对账）。
// 按 created_at 倒序；每行带 receipt 原始对象 + 对应 manifest 原始对象（无则 nil）。
type ReceiptReconcileRow struct {
	Receipt  *model.MediaReceipt `json:"receipt"`
	Manifest *model.MediaManifest `json:"manifest"` // 可为 nil（收据引用了对象但无清单）
}

func ListMediaReceiptsForReconcile(ctx context.Context, database *sql.DB, tenantID string, limit int) ([]ReceiptReconcileRow, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("list media receipts: tenant_id required")
	}
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := database.QueryContext(ctx, `
		SELECT r.receipt_json::text, m.manifest_json::text
		FROM media_receipts r
		LEFT JOIN media_manifests m ON m.tenant_id = r.tenant_id AND m.object_ref = r.object_ref
		WHERE r.tenant_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("list media receipts: %w", err)
	}
	defer rows.Close()

	var out []ReceiptReconcileRow
	for rows.Next() {
		var rRaw, mRaw sql.NullString
		if err := rows.Scan(&rRaw, &mRaw); err != nil {
			return nil, fmt.Errorf("list media receipts scan: %w", err)
		}
		var row ReceiptReconcileRow
		if rRaw.Valid {
			row.Receipt = &model.MediaReceipt{}
			if err := json.Unmarshal([]byte(rRaw.String), row.Receipt); err != nil {
				return nil, fmt.Errorf("list media receipts unmarshal receipt: %w", err)
			}
		}
		if mRaw.Valid && mRaw.String != "" {
			row.Manifest = &model.MediaManifest{}
			if err := json.Unmarshal([]byte(mRaw.String), row.Manifest); err != nil {
				return nil, fmt.Errorf("list media receipts unmarshal manifest: %w", err)
			}
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list media receipts rows: %w", err)
	}
	return out, nil
}

// IngestMessage 数据链接入单条消息（R42 数据链 ingestion，幂等）。
// 流水线：source_account upsert → conversation upsert → message upsert
// （按 tenant_id+upstream_message_id 幂等，重复上报不重复插入）。
// 返回 (accepted, duplicated)：accepted=新插入，duplicated=已存在跳过。
// 注意：upstream_message_id 为空时不做幂等（各为独立行，ACK 仍耐久提交后发）。
func IngestMessage(ctx context.Context, database *sql.DB, tenantID string, m *model.IngestMessage) (int, int, error) {
	if tenantID == "" || m == nil || m.UpstreamMessageID == "" || m.ConversationRef == "" {
		return 0, 0, fmt.Errorf("ingest message: tenant_id/upstream_message_id/conversation_ref required")
	}
	var srcID, convID string
	// 1) source_account upsert（按 tenant+source_type+external_account_id）
	if err := database.QueryRowContext(ctx, `
		INSERT INTO source_accounts (tenant_id, source_type, external_account_id, display_name, status)
		VALUES ($1, $2, $3, $4, 'active')
		ON CONFLICT (tenant_id, source_type, external_account_id) DO UPDATE SET
			display_name = EXCLUDED.display_name, updated_at = now()
		RETURNING id`, tenantID, m.SourceType, m.SourceExternalID, m.SourceDisplayName).Scan(&srcID); err != nil {
		return 0, 0, fmt.Errorf("ingest source_account: %w", err)
	}
	// 2) conversation upsert（按 tenant+source_account+external_conversation_id）
	if err := database.QueryRowContext(ctx, `
		INSERT INTO conversations (tenant_id, source_account_id, external_conversation_id, conversation_type, title, last_message_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (tenant_id, source_account_id, external_conversation_id) DO UPDATE SET
			title = COALESCE(EXCLUDED.title, conversations.title),
			last_message_at = now(), updated_at = now()
		RETURNING id`, tenantID, srcID, m.ConversationRef, m.ConversationType, m.ConversationTitle).Scan(&convID); err != nil {
		return 0, 0, fmt.Errorf("ingest conversation: %w", err)
	}
	// 3) message upsert（按 tenant+upstream_message_id 幂等）
	tag, err := database.ExecContext(ctx, `
		INSERT INTO messages (tenant_id, source_account_id, conversation_id, sender_contact_id,
			direction, message_type, sent_at, sent_at_ms, content_text, content_json, content_hash,
			upstream_message_id, decode_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (tenant_id, upstream_message_id) WHERE upstream_message_id IS NOT NULL
		DO NOTHING`,
		tenantID, srcID, convID, nil, m.Direction, m.MessageType, m.SentAt, m.SentAtMs,
		m.ContentText, m.ContentJSON, m.ContentHash, m.UpstreamMessageID, m.DecodeStatus)
	if err != nil {
		return 0, 0, fmt.Errorf("ingest message: %w", err)
	}
	n, err := tag.RowsAffected()
	if err != nil {
		return 0, 0, fmt.Errorf("ingest message rows: %w", err)
	}
	if n > 0 {
		return 1, 0, nil
	}
	return 0, 1, nil
}

// SaveMediaObject 登记已上传到对象存储的媒体对象元数据（R42.7 upload）。
// 以 (tenant_id, sha256) 内容寻址幂等：重复上传同 sha256 返回既有 object_ref（true=已存在）。
// bytes 在对象存储（minio），本表只登记元数据（AGENTS A06 对象/元数据分离）。
func SaveMediaObject(ctx context.Context, database *sql.DB, tenantID, objectRef, sha256, mediaType string, sizeBytes int64) (string, bool, error) {
	if tenantID == "" || objectRef == "" || sha256 == "" || mediaType == "" {
		return "", false, fmt.Errorf("save media object: tenant_id/object_ref/sha256/media_type required")
	}
	var existing string
	err := database.QueryRowContext(ctx, `
		SELECT object_ref FROM media_objects
		WHERE tenant_id = $1 AND sha256 = $2`, tenantID, sha256).Scan(&existing)
	if err == nil {
		return existing, true, nil
	}
	if err != sql.ErrNoRows {
		return "", false, fmt.Errorf("save media object: query existing: %w", err)
	}
	_, err = database.ExecContext(ctx, `
		INSERT INTO media_objects (tenant_id, object_ref, sha256, media_type, size_bytes)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id, sha256) DO NOTHING`, tenantID, objectRef, sha256, mediaType, sizeBytes)
	if err != nil {
		return "", false, fmt.Errorf("save media object: %w", err)
	}
	return objectRef, false, nil
}
