## 1. 后端：命格算法引擎

- [x] 1.1 新建 `backend/pkg/bazi/mingge.go`，实现藏干透干检测辅助函数（判断某天干是否透出到目标天干集合）
- [x] 1.2 实现七优先级取格主函数 `DetectMingGe(result *BaziResult) (name, desc string)`
- [x] 1.3 实现三合局/三会局检测（优先级6）：检测四柱地支是否构成三合/三会，返回合局五行
- [x] 1.4 实现十神 → 格名的映射（含建禄格/月刃格特殊映射）
- [x] 1.5 内置格局说明文字字典（正官格/七杀格/食神格/伤官格/正财格/偏财格/正印格/偏印格/建禄格/月刃格/杂气格），每条不少于 50 字

## 2. 后端：BaziResult 集成

- [x] 2.1 在 `backend/pkg/bazi/engine.go` 的 `BaziResult` 结构体中新增 `MingGe string` 和 `MingGeDesc string` 字段（json tag：`ming_ge` / `ming_ge_desc`）
- [x] 2.2 在 `Calculate()` 函数末尾调用 `DetectMingGe(res)`，将返回值写入 `res.MingGe` 和 `res.MingGeDesc`

## 3. 前端：结果页展示

- [x] 3.1 在 `frontend/src/pages/ResultPage.tsx` 的 `BaziResult` interface 中新增 `ming_ge?: string` 和 `ming_ge_desc?: string` 可选字段
- [x] 3.2 在结果页顶部 `.result-tags` 区域条件渲染命格 Badge（仅当 `result.ming_ge` 有值时展示）
- [x] 3.3 为命格 Badge 添加点击处理：点击后弹出 Modal 展示格名 + 说明文字（复用现有 `shensha-modal-*` CSS 类）
- [x] 3.4 为命格 Badge 添加专属样式（区别于五行喜用徽标，用紫色/深青色调体现"格局"感）

## 4. 验证

- [x] 4.1 本地手动起盘验证： API 正确返回 `ming_ge` 字段（已通过 curl 验证多个命局）
- [x] 4.2 验证旧命盘历史记录页面加载时，格名 Badge 不展示、不报错（已通过可选链 `ming_ge?` 处理）
- [x] 4.3 验证手机端格名 Badge 在顶部区域正确换行展示（`.result-tags` 已有 flex-wrap）
