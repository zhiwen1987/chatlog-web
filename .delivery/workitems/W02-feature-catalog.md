# W02 串行共享窗口：feature-catalog 冻结

状态: 完成（骨架 + 结构校验，不含类型生成）
owner: 协调者（Lead）串行窗口
base: HEAD 067c023（未提交改动保留原样）

## 改动（全部为只增，未覆盖/删除现有文件）
- 新建 packages/contracts/feature-catalog.json
  - 23 个功能 key（R42.2 全表）、catalog_version=1、rules 语义声明
  - 7 条依赖（media.upload.* → archive.ingest；media.transcode →
    media.preview；ai.profile → customer.manage；mcp.write → mcp.read）
  - DAG 无环，唯一 key，无通配符/all=true
- 新建 scripts/verify_feature_catalog.py
  - 结构校验：key 唯一/非空/无空白/无通配符、依赖引用存在、DAG 无环、
    grant 窗口左闭右开、quotas 为对象
- 更新 packages/contracts/README.md（占位 → 登记 feature-catalog + 待冻结清单）

## 证据
- `python3 scripts/verify_feature_catalog.py packages/contracts/feature-catalog.json`
  → OK: 23 features, unique keys, acyclic DAG
- 坏例反例：追加 `no.such.key` 依赖 → FAIL 2 处（dangling dep + cycle），exit=1
  （校验器不是摆设）

## 未做（属于后续工作项，本项为合同骨架）
- 未生成 Go/TypeScript 类型、平台表单、产品状态映射（R42.2 要求由目录生成）
- 未做 JWS 签名/时间/配额/绑定语义校验（R42.8，需真实签发与验签代码）
- 未更新 MANIFEST.sha256（保持既有范围差异口径，待用户处理 59 处未提交内容后统一）

## 下一步
- W03 合同扩展：data-ownership / license claims schema / IntegrityReport 骨架
- 或 W02b 生成器：Go/TS 类型生成 + 平台表单映射（R42.2）