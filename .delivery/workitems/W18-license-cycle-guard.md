# W18 串行共享窗口：license dependencyCheck 循环依赖防护

状态: 完成（测试驱动发现真实缺陷，修复+回归）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改文件未触碰

## 问题发现（测试驱动）
实测 `dependencyCheck` 遇依赖环（A→B→A）时抛
`RangeError: Maximum call stack size exceeded`——无限递归。
feature-catalog.json 虽保证 DAG 无环，但运行时可能拿到损坏目录，须防御。

## 修复
- `apps/admin-web/src/lib/license.js` 的 `dependencyCheck`
  - 加 visiting 集合检测循环：已访问节点视为依赖不满足（missing 记录），
    防无限递归；正常无环依赖仍按原有语义求值
- `apps/admin-web/tests/license.test.mjs` 补循环依赖测试
  - 断言：available=false 且循环节点在 missing 中，不抛异常

## 证据
- 修复前实测：`THREW: RangeError Maximum call stack size exceeded`
- 修复后实测：`cycle result: {"available":false,"missing":["a"]} (no throw)`
  + 正常场景 `{"available":true,"missing":[]}` 回归通过
- `node --test tests/license.test.mjs` → 7/7；全量 `tests/*.test.mjs` → 80/80
- 5 个 Python 校验器 + A14 全绿
- 用户 11 处已修改文件未触碰

## 没有做/限制
- 未改 feature-catalog.json 的 DAG 保证（它是源，仍无环）；防护是运行时防御
- 未更新 MANIFEST.sha256（保持既有口径）

## 下一项
- 串行窗口测试加固继续：media 对账/ownership 大数边界等可补测试
- P1 跨层真实编码仍等用户决策
