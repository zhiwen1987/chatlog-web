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
