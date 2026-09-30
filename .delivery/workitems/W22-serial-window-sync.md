# W22 串行共享窗口：综合终检覆盖补全（source.js 纳入门禁）

状态: 完成（纯新增，综合终检通过）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改文件未触碰

## 背景
W17 的 verify_serial_window.py 前端产物列表只含 5 个工具
（integrity/license/media/ownership/contracts），未纳入 W21 新增的 source.js。
综合终检门禁落后于成果，须同步。

## 改动（只增）
- 更新 `scripts/verify_serial_window.py`
  - FRONTEND_LIBS 加 `apps/admin-web/src/lib/source.js`
  - FRONTEND_TESTS 加 `apps/admin-web/tests/source.test.mjs`

## 证据
- 更新前：终检脚本列表不含 source.js（覆盖缺口）
- 更新后：`python3 scripts/verify_serial_window.py` → OK, exit=0
- 全量：87 node 测试 + 6 校验器 + A14 全绿
- 用户 11 处已修改文件未触碰

## 没有做/限制
- 未更新 MANIFEST.sha256（保持既有口径）

## 下一项
- 串行窗口门禁与成果同步完成（W02-W22）
- P1 跨层真实编码仍等用户决策
