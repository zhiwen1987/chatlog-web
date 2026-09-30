# W02b 串行共享窗口：feature 类型生成器（Go/TS → 实际 Go/JS）

状态: 完成（生成器 + 幂等校验 + 反证测试；Go 编译未测因环境无工具链）
owner: 协调者（Lead）串行窗口
base: 用户 59 处未提交内容保留原样

## 改动（全部为只增，未覆盖/删除现有文件）
- 新建 `scripts/gen_feature_types.py`
  - 从 `packages/contracts/feature-catalog.json`（唯一源）生成两类产物
  - Go：`apps/server/internal/model/features.go`
    - `Feature*` 常量（如 FeatureArchiveRead = "archive.read"）
    - `FeatureDeps` map[string][]string 依赖表
  - JS：`apps/admin-web/src/lib/features.js`
    - `FEATURES` 常量对象
    - `FEATURE_CATALOG` 平台表单映射（name/description/dependsOn）
  - `--check` 幂等守护：产物与源不一致即失败（防 app 目录手写副本漂移，R42.2）
- 命名修复：`media.document_extract` → `FeatureMediaDocumentExtract`（按 '_' 拆分驼峰）

## 证据
- `python3 scripts/gen_feature_types.py` → 生成两个产物文件
- `python3 scripts/gen_feature_types.py --check` → OK: feature types in sync
- 反证：向源临时追加 `x.new.feature` → `--check` FAIL（两个文件失步，exit=1）
  → 还原源 → `--check` 恢复 OK（幂等守护真实生效）
- Python 语法级校验：Go 23 常量 key 全部与源一致、无未知引用；
  JS 23 键与源完全一致

## 没有做/限制（如实）
- `go vet` / `go build` 未执行：本机无 Go 工具链（`go: command not found`，
  无 /usr/local/go、/opt/homebrew/go）
- JS 侧未跑 lint/test（admin-web 需 npm install 后才有入口，未碰依赖）
- 未更新 MANIFEST.sha256（保持既有范围差异口径）
- 生成器不含平台表单的复杂映射规则（如授权/ACL/UI 状态机），只做 R42.2 目录→常量/映射基础

## 下一项
- 按 P1.5 合同扩展线已收口；P1 真实业务编码等待用户处理 59 处未提交内容解锁 worktree 并行
- 或先做串行只读复核/文档同步（A14：行为变更后更新功能说明/接口文档）