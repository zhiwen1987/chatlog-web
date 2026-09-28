package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	// 确保环境变量隔离
	t.Setenv("CHATLOG_ADDR", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("CHATLOG_JWT_SECRET", "")
	cfg := Load()
	if cfg.Addr != ":8080" {
		t.Fatalf("expected default addr :8080, got %q", cfg.Addr)
	}
	if cfg.JWTSecret != "dev-secret-change-me" {
		t.Fatalf("expected default jwt secret, got %q", cfg.JWTSecret)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("CHATLOG_ADDR", ":9090")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("CHATLOG_JWT_SECRET", "custom")
	t.Setenv("CHATLOG_TOKEN_TTL_MINUTES", "120")
	cfg := Load()
	if cfg.Addr != ":9090" {
		t.Fatalf("addr: got %q", cfg.Addr)
	}
	if cfg.DatabaseURL != "postgres://x" {
		t.Fatalf("database url: got %q", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "custom" {
		t.Fatalf("jwt secret: got %q", cfg.JWTSecret)
	}
	if cfg.TokenTTLMinutes != 120 {
		t.Fatalf("ttl: got %d", cfg.TokenTTLMinutes)
	}
}