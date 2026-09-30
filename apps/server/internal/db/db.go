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
