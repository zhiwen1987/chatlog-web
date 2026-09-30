# W08 串行共享窗口：data-ownership 前端校验工具（admin-web src/lib）

状态: 完成（纯新增 JS + node:test 真实通过）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改 + 59 处未跟踪未触碰（本轮未触碰任何既有文件）

## 改动（全部为只增）
- 新建 `apps/admin-web/src/lib/ownership.js`
  - `decimalStringError(value, label)`：精确十进制字符串校验（R42.7 field_mode）
    64位ID/bytes/seq/revision 必须用十进制字符串，不用浮点；浮点/负数/
    科学计数/非数字拒绝；undefined/null/'' 保持缺失（不补 0）
  - `checkDecimalFields(obj, fields)`：批量校验字段的十进制表示
  - `timeUnknownError(value)`：未知时间不得变成今天/0/1970；
    接受 null/'unknown'，拒绝 '' 被伪装
  - `ownerReport(rows, knownOwners)`：数据归属域登记检查
    列出 data-ownership.json 未登记的 owner（防数据归属逃逸）
- 新建 `apps/admin-web/tests/ownership.test.mjs`
  - node:test + data:text/javascript 动态加载

## 证据
- `node --test tests/ownership.test.mjs` → pass 6 / fail 0
- 全量回归 `node --test tests/*.test.mjs` → tests 73 / pass 73 / fail 0
  （新增 ownership 6 + 前四轮新增 22 + 既有 45）
- 反证：`9007199254740993` 数字字面量（超 int53）被拒、`'1.5'/'1e10'/'abc'` 被拒、
  未登记 owner 被标记

## 没有做/限制（如实）
- 未接 UI 视图（未在任何 .vue 挂载；视图为已修改文件，不触碰）
- 未做跨 tenant FK/分区唯一/RLS 运行时检查（R07/R39 属 DB 层，前端不做）
- 未更新 MANIFEST.sha256（保持既有范围差异口径）
- data-ownership.json 的 writes/io/retention 未做前端强校验（面向发行方/后台，非前端消费）

## 下一项
- 串行纵切已连做 4 个前端消费工具（W05/W06/W07/W08），`npm test` 全量 73 全绿。
  纯新增前端纵切已到合理边界；再往下需触碰用户已修改 .vue 或跨层 Go/Rust
  （无工具链/需 worktree 隔离）
- P1 跨层真实编码仍等用户处理 59 处未提交内容解锁 worktree 并行