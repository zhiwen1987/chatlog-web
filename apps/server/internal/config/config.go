package config

import (
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
}

// Load 从环境变量读取配置。
func Load() Config {
	return Config{
		Addr:            getEnv("CHATLOG_ADDR", ":8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://chatlog:chatlog@localhost:5432/chatlog?sslmode=disable"),
		JWTSecret:       getEnv("CHATLOG_JWT_SECRET", "dev-secret-change-me"),
		TokenTTLMinutes: getEnvInt("CHATLOG_TOKEN_TTL_MINUTES", 60*24*7),
		AuditEnabled:    getEnvBool("CHATLOG_AUDIT_ENABLED", true),
	}
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