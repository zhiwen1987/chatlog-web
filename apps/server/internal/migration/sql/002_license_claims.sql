-- 002_license_claims.sql
-- 许可 claims v2 持久化（R42.8）。
-- 单一 claims JSON 行（licensee+deployment 复合唯一）；feature_grants 存 JSONB，
-- 保持 license-claims.json 合同为唯一结构源，不拆多表复制 grant 字段。
-- 仅发行方后台写入；发证私钥不进入产品/源码/日志（AGENTS）。

CREATE TABLE IF NOT EXISTS license_claims (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  licensee      TEXT NOT NULL,
  deployment_id TEXT NOT NULL,
  claims_json   JSONB NOT NULL,
  license_revision INT NOT NULL DEFAULT 1,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (licensee, deployment_id)
);

CREATE INDEX IF NOT EXISTS idx_license_claims_deployment
  ON license_claims (deployment_id);

-- 只读查询角色访问（应用层读取）；写入仅发行方后台/授权迁移。
ALTER TABLE license_claims ENABLE ROW LEVEL SECURITY;
CREATE POLICY license_claims_read ON license_claims
  FOR SELECT USING (true);