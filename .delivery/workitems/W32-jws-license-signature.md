# W32 跨层：JWS license 验签/签发闭环（R42.8 签名缺口）

状态: 完成（auth/model/handler 真实 PG 全绿）
owner: 协调者（Lead）串行+全自动推进
base: main 340cb5c（工作树干净）

## 改动
- `apps/server/internal/auth/claims_jws.go`
  - `SignClaims`：HS256 对称 + RS256 非对称（显式算法绑定防混淆）
  - `VerifyClaimsJWS`：验签通过返回 `*model.Claims`；篡改/错钥/错aud → `ErrLicenseSignature`，
    过期 → `ErrLicenseExpired`；签发方私钥不进入产品
- `apps/server/internal/auth/claims_jws_test.go`：7 单测全 PASS
  - 正验 / 篡改 / 错钥 / 过期 / 错aud / RS256 正验 / RS256 错钥
  - 全自生成测试密钥，不依赖真实签发方

## 证据
- `go vet ./...` exit=0
- auth/model/handler 真实 PG（chatlog-postgres chatlog_test）全绿
- 提交 `2854700`，工作树干净

## 没有做/限制
- **未接入 `LoadLicenseClaims`**：验签逻辑独立成层，尚未在生产加载路径接线
  （接入需配置验签公钥来源，属部署/密钥注入边界，非本工作项）
- 未生成真实签发方私钥/证书链（属发证域，需独立 Issuer 环境）

## 下一项
- 接入 LoadLicenseClaims 验签路径（需部署密钥配置）
- 服务端镜像重建让 license/device/heartbeat 端点在运行中服务上线（需授权动容器）