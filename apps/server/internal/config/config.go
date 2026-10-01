package config

import (
	"encoding/base64"
	"os"
	"strconv"
)

// Config 服务端配置，全部来自环境变量（Docker Compose 注入）。
type Config struct {
	Addr            string
	DatabaseURL     string
	JWTSecret       string
	TokenTTLMinutes int
	AuditEnabled    bool
	DeploymentID    string

	// LicenseVerifyKey 可选：license claims JWS 验签对称密钥（base64 编码）。
	// 配置后加载路径走 JWS 验签；未配置保持直接解析 JSON（兼容现状）。
	LicenseVerifyKey []byte
	// LicenseExpectedAud 可选：验签要求的 audience（部署域绑定）。
	LicenseExpectedAud string

	// MinioEndpoint/MinioAccessKey/MinioSecretKey/MinioBucket 可选：媒体对象存储（upload bytes 级）。
	// 配置 MINIO_ENDPOINT 后启用 upload 端点；未配置时端点返回 503（对象存储未接入）。
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
}

// Load 从环境变量读取配置。
func Load() Config {
	c := Config{
		Addr:               getEnv("CHATLOG_ADDR", ":8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://chatlog:chatlog@localhost:5432/chatlog?sslmode=disable"),
		JWTSecret:          getEnv("CHATLOG_JWT_SECRET", "dev-secret-change-me"),
		TokenTTLMinutes:    getEnvInt("CHATLOG_TOKEN_TTL_MINUTES", 60*24*7),
		AuditEnabled:       getEnvBool("CHATLOG_AUDIT_ENABLED", true),
		DeploymentID:       getEnv("CHATLOG_DEPLOYMENT_ID", "dev"),
		LicenseExpectedAud: getEnv("CHATLOG_LICENSE_AUDIENCE", ""),
	}
	if k := getEnv("CHATLOG_LICENSE_VERIFY_KEY", ""); k != "" {
		if raw, err := base64.StdEncoding.DecodeString(k); err == nil {
			c.LicenseVerifyKey = raw
		}
	}
	c.MinioEndpoint = getEnv("MINIO_ENDPOINT", "")
	c.MinioAccessKey = getEnv("MINIO_ACCESS_KEY", "")
	c.MinioSecretKey = getEnv("MINIO_SECRET_KEY", "")
	c.MinioBucket = getEnv("MINIO_BUCKET", "chatlog-media")
	c.MinioUseSSL = getEnvBool("MINIO_USE_SSL", false)
	return c
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
