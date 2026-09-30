package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

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
