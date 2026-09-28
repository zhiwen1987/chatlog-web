package auth

import (
	"testing"
	"time"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("s3cret-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "s3cret-password" {
		t.Fatal("password stored in plaintext")
	}
	if !VerifyPassword(hash, "s3cret-password") {
		t.Fatal("VerifyPassword should accept correct password")
	}
	if VerifyPassword(hash, "wrong") {
		t.Fatal("VerifyPassword should reject wrong password")
	}
}

func TestParseBearer(t *testing.T) {
	tok, err := ParseBearer("Bearer abc123")
	if err != nil || tok != "abc123" {
		t.Fatalf("expected abc123, got %q err %v", tok, err)
	}
	if _, err := ParseBearer("Basic abc"); err != ErrMalformedBearer {
		t.Fatalf("expected ErrMalformedBearer, got %v", err)
	}
	if _, err := ParseBearer(""); err != ErrMalformedBearer {
		t.Fatalf("expected ErrMalformedBearer for empty, got %v", err)
	}
}

func TestNewToken(t *testing.T) {
	a, _ := NewToken()
	b, _ := NewToken()
	if a == "" || a == b {
		t.Fatalf("expected distinct non-empty tokens, got %q %q", a, b)
	}
}

func TestJWT(t *testing.T) {
	secret := "test-secret"
	tok, err := SignJWT(secret, "user-1", "tenant-1", "owner", time.Hour)
	if err != nil {
		t.Fatalf("SignJWT: %v", err)
	}
	claims, err := ParseJWT(secret, tok)
	if err != nil {
		t.Fatalf("ParseJWT: %v", err)
	}
	if claims.UserID != "user-1" || claims.TenantID != "tenant-1" || claims.Role != "owner" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	// 错误签名
	if _, err := ParseJWT("wrong-secret", tok); err == nil {
		t.Fatal("expected error for wrong secret")
	}
	// 过期 token
	exp, _ := SignJWT(secret, "u", "t", "owner", -time.Minute)
	if _, err := ParseJWT(secret, exp); err == nil {
		t.Fatal("expected expired error")
	}
}