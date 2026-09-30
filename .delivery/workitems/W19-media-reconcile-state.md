# W19 串行共享窗口：media reconcile 状态语义修复

状态: 完成（测试驱动发现真实语义缺陷，修复+回归）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改文件未触碰

## 问题发现（测试驱动）
实测 `reconcile` 对 `deleting`/`deleted` 状态返回 ok=true——语义错误：
receipt 是"已持久化可接收"的证明，对象已删/删除中不应 ACK。
原实现只拒绝 `quarantined`，漏了 `deleting`/`deleted`。

## 修复
- `apps/admin-web/src/lib/media.js` 的 `reconcile`
  - 仅 `stored` 状态可确认接收；`deleting`/`deleted`/`quarantined` 均拒绝
    （返回 `对象处于 <state>，不能确认接收`）
- `apps/admin-web/tests/media.test.mjs` 补 deleting/deleted 拒绝 + stored 通过测试

## 证据
- 修复前实测：`stored->ok:true / deleting->ok:true / deleted->ok:true / quarantined->ok:false`
- 修复后实测：仅 stored ok:true，其余拒绝（见测试）
- `media.test.mjs` → 9/9；全量 `tests/*.test.mjs` → 81/81
- 5 个 Python 校验器 + A14 全绿
- 用户 11 处已修改文件未触碰

## 没有做/限制
- 未改 protocol 的 media-manifest schema（state 枚举不变；语义由消费端强化）
- 未更新 MANIFEST.sha256（保持既有口径）

## 下一项
- 串行窗口测试加固持续：可继续补 ownership/integrity 边界测试
- P1 跨层真实编码仍等用户决策
