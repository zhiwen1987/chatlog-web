package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 生成 bcrypt 密码哈希。
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// VerifyPassword 校验密码（常量时间比较哈希结果）。
func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// NewToken 生成随机 256 位令牌（用于设备长期凭证 / 邀请链接等）。
func NewToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

var ErrMalformedBearer = errors.New("malformed bearer token")

// ParseBearer 从 Authorization 头解析 Bearer token。
func ParseBearer(header string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", ErrMalformedBearer
	}
	tok := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if tok == "" || strings.ContainsAny(tok, " \t\r\n") {
		return "", ErrMalformedBearer
	}
	return tok, nil
}

// SecureEquals 常量时间字符串比较。
func SecureEquals(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}