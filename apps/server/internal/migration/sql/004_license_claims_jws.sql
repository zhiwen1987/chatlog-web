-- 004_license_claims_jws.sql
-- license claims JWS 签名列（R42.8 签名要求）。
-- 发行方签发时把紧凑 JWS 字符串写入 claims_jws；产品验签路径读取该列。
-- 未签名的旧行保持 claims_json 直接解析兼容（未配置验签密钥的现状）。
ALTER TABLE license_claims ADD COLUMN IF NOT EXISTS claims_jws TEXT;