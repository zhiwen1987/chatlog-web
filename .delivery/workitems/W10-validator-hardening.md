# W10 串行共享窗口：校验器加固（required 一致性检查）

状态: 完成（反证驱动加固，正例+坏例全验证）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改 + 61 处未跟踪未触碰（只改 2 个 Python 校验器脚本，未碰 schema/代码/文档）

## 问题发现（反证驱动）
复验 W03/W04 坏例时暴露真实漏洞：删除 schema 的 `required` 数组里某必填字段后，
校验器仍报 OK——因为原检查只验证"properties 里的字段是否在 required 列表中"，
而忽略"required 列表是否与声明的必填字段一致"。

## 修复（2 个 Python 校验器）
- `scripts/verify_contract_schemas.py`
  - license-claims、integrity-report：required 数组与声明的必填字段列表 set 相等检查
- `scripts/verify_protocol_schemas.py`
  - media-manifest、media-receipt：同上
- 保留原有"required 字段须有 property 定义"检查（方向不变）

## 证据
- 正例：4 个校验器全过（contract-schemas/protocol-schemas/feature-catalog/gen_feature_types --check）
- 坏例（临时改坏 required 后还原）：
  - 删 license-claims required 的 typ → FAIL required 不一致，exit=1
  - 删 media-receipt required 的 backup_set → FAIL required 不一致，exit=1
  - 恢复后全通过
- 全量回归：node 73/73、A14 exit=0、Python 4 校验器全绿
- 用户 11 处已修改 + 61 处未跟踪未触碰

## 没有做/限制（如实）
- 未做 JSON-Schema 完整校验（如 jsonschema 库对全部字段类型的验证，当前用轻量结构检查）
- 未更新 MANIFEST.sha256（保持既有范围差异口径）

## 下一项
- 串行安全窗口工作已全部收口（W02-W10）
- P1 跨层真实编码仍等用户处理 61 处未提交内容解锁 worktree 并行，或授权主工作树直改