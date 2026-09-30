package auth

// claims_jws.go — license-claims v2 的 JWS 签发与验签（R42.8）。
// 签发方侧用私钥签名（HS256 对称或 RS256 非对称），产品侧用公钥/对称密钥验签。
// 验签通过后返回 *model.Claims 供授权判定；验签失败返回 ErrLicenseSignature。
// 发证私钥只属于签发方域，不进入产品/源码/日志（AGENTS A07）。

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

var (
	// ErrLicenseSignature 验签失败（签名无效/密钥不匹配/篡改/过期）。
	ErrLicenseSignature = errors.New("license signature invalid")
	// ErrLicenseExpired 已过期（验签算法接受但时间窗失效）。
	ErrLicenseExpired = errors.New("license expired")
)

// signMethod 决定签名算法：非空 RSA 私钥→RS256，否则 HS256。
// key 必须非 nil；对称密钥长度 ≥32 字节（HS256 安全下限），否则返回错误。
func signMethod(key any) (jwt.SigningMethod, error) {
	switch k := key.(type) {
	case []byte:
		if len(k) < 32 {
			return nil, fmt.Errorf("hmac key too short: %d bytes", len(k))
		}
		return jwt.SigningMethodHS256, nil
	case *rsa.PrivateKey:
		return jwt.SigningMethodRS256, nil
	default:
		return nil, fmt.Errorf("unsupported signing key type %T", key)
	}
}

// ClaimsJWSPayload 与 model.Claims 同构但含时间字段，用于签发/验签的原始 payload。
// 注意：model.Claims 是"授权判定用"精简结构；签发/验签需保留 iat/exp/iss/aud 等注册字段，
// 故这里用 jwt.RegisteredClaims + 业务字段。
type ClaimsJWSPayload struct {
	Typ               string          `json:"typ"`
	Licensee          string          `json:"licensee"`
	Deployment        string          `json:"deployment"`
	FeatureCatalogVer int             `json:"feature_catalog_version"`
	LicenseRevision   int             `json:"license_revision"`
	FeatureGrants     []model.Grant   `json:"feature_grants"`
	jwt.RegisteredClaims
}

// SignClaims 用签发方密钥签发 JWS。key 为 []byte（HS256）或 *rsa.PrivateKey（RS256）。
// 返回紧凑 JWS 字符串。
func SignClaims(key any, p *ClaimsJWSPayload) (string, error) {
	method, err := signMethod(key)
	if err != nil {
		return "", err
	}
	tok := jwt.NewWithClaims(method, p)
	return tok.SignedString(key)
}

// VerifyClaimsJWS 验签并解析 license claims。verificationKey 为 []byte（HS256）或 *rsa.PublicKey（RS256）。
// 返回解析后的 *model.Claims（不含未注册字段的校验——签名本身保证完整性）。
// 过期返回 ErrLicenseExpired；签名无效/算法不符返回 ErrLicenseSignature。
func VerifyClaimsJWS(verificationKey any, tokenString string, expectedAud string) (*model.Claims, error) {
	if verificationKey == nil {
		return nil, ErrLicenseSignature
	}
	tok, err := jwt.ParseWithClaims(tokenString, &ClaimsJWSPayload{}, func(t *jwt.Token) (any, error) {
		// 限定算法族，防算法混淆（HS256/RS256 显式匹配密钥类型）。
		switch verificationKey.(type) {
		case []byte:
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
		case *rsa.PublicKey:
			if t.Method != jwt.SigningMethodRS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
		default:
			return nil, fmt.Errorf("unsupported verification key type %T", verificationKey)
		}
		return verificationKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrLicenseExpired
		}
		return nil, ErrLicenseSignature
	}
	p, ok := tok.Claims.(*ClaimsJWSPayload)
	if !ok || !tok.Valid {
		return nil, ErrLicenseSignature
	}
	// 校验 audience（部署域绑定）。
	if expectedAud != "" {
		found := false
		for _, a := range p.Audience {
			if a == expectedAud {
				found = true
				break
			}
		}
		if !found {
			return nil, ErrLicenseSignature
		}
	}
	// 过期与生效时间（iat 未来视为无效签名）。
	now := time.Now()
	if p.ExpiresAt != nil && p.ExpiresAt.Before(now) {
		return nil, ErrLicenseExpired
	}
	if p.NotBefore != nil && p.NotBefore.After(now) {
		return nil, ErrLicenseSignature
	}
	if p.IssuedAt != nil && p.IssuedAt.After(now.Add(5*time.Minute)) {
		return nil, ErrLicenseSignature
	}
	return &model.Claims{
		Typ:               p.Typ,
		Iss:               p.Issuer,
		Aud:               strings.Join(p.Audience, ","),
		Licensee:          p.Licensee,
		Deployment:        p.Deployment,
		FeatureCatalogVer: p.FeatureCatalogVer,
		LicenseRevision:   p.LicenseRevision,
		FeatureGrants:     p.FeatureGrants,
	}, nil
}