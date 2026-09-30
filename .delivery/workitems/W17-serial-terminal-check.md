# W17 串行共享窗口：综合终检门禁（verify_serial_window.py）

状态: 完成（纯新增，全量验证通过）
owner: 协调者（Lead）串行窗口
base: 用户 11 处已修改文件未触碰

## 背景
串行窗口 16 项工作全部收口后，建一个可重复运行的综合验收门禁，
聚合所有产物存在性 + 校验器 + 前端测试，作为 A14 收尾的最终验收。

## 改动（只增）
- 新建 `scripts/verify_serial_window.py`
  - 校验 8 个 schema 文件存在（contracts/protocol/source-contract/profile-schema）
  - 校验 5 个 Python 校验器 + gen --check + A14 复核通过
  - 校验前端 5 个工具 + 5 个测试文件存在
  - 在 apps/admin-web 目录真实执行 `node --test tests/*.test.mjs`（79 个）
  - 退出码 0=全部通过

## 证据
- `python3 scripts/verify_serial_window.py` → OK, exit=0
- 调试修正：node 测试执行目录原为仓库根（错误），改为 apps/admin-web 后真实执行
  （confirmed: tests 79 / pass 79 / fail 0）
- 用户 11 处已修改文件未触碰

## 没有做/限制
- 未更新 MANIFEST.sha256（保持既有口径）
- 门禁不替代真实产品验收（PG/发证/TB 等），只验收串行窗口合同与前端产物

## 下一项
- 串行窗口全部工作收口（W02-W17 + S01）
- P1 跨层真实编码（SourceAdapter/AI 画像/服务端 API/心跳/ingestion）仍等用户决策
