package auth

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey int

const (
	ctxKeyUserID ctxKey = iota
	ctxKeyTenantID
	ctxKeyRole
	ctxKeyDeviceID
)

// ContextUserID 从请求上下文读取用户 ID。
func ContextUserID(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyUserID).(string)
	return v
}

// ContextTenantID 从请求上下文读取租户 ID。
func ContextTenantID(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyTenantID).(string)
	return v
}

// ContextRole 从请求上下文读取角色。
func ContextRole(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyRole).(string)
	return v
}

// ContextDeviceID 从请求上下文读取设备 ID。
func ContextDeviceID(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyDeviceID).(string)
	return v
}

// Authenticate 解析 Bearer JWT 并注入上下文。
func Authenticate(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := ParseBearer(r.Header.Get("Authorization"))
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		claims, err := ParseJWT(secret, token)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ctxKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ctxKeyTenantID, claims.TenantID)
		ctx = context.WithValue(ctx, ctxKeyRole, claims.Role)
		ctx = context.WithValue(ctx, ctxKeyDeviceID, claims.DeviceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRoles 校验角色白名单；角色为空表示任意已认证用户。
func RequireRoles(allowed ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := ContextRole(r.Context())
			if role == "" {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}
			if len(allowed) == 0 {
				next.ServeHTTP(w, r)
				return
			}
			for _, a := range allowed {
				if strings.EqualFold(a, role) {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		})
	}
}