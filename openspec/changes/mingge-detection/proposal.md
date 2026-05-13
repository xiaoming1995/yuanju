## Why

命盘结果页目前展示了四柱、五行、用神忌神等详细信息，但缺少**命格（格局）**这一核心定性标签。命格是传统八字命理中统领全局的关键概念，决定大运流年吉凶判断的基准框架，专业用户和普通用户都期望在命盘中第一眼看到它。

## What Changes

- **新增后端命格推算引擎**：实现七优先级透干取格算法（`pkg/bazi/mingge.go`），涵盖月令正格及兜底逻辑
- **扩展 BaziResult 结构体**：新增 `ming_ge`（格名）、`ming_ge_desc`（格局说明文字）两个字段
- **前端命盘顶部展示**：在结果页标题区的喜用/忌 Badge 行追加命格 Badge `[正官格]`
- **点击弹出 Modal 说明**：复用现有神煞注解 Modal 组件，展示格名 + 格局简介文字

## Capabilities

### New Capabilities

- `mingge-detection`: 基于七优先级透干取格算法，计算命盘所属格局（命格），并在前端结果页顶部展示格名 Badge，点击弹出格局说明 Modal

### Modified Capabilities

- `bazi-advanced-data`: BaziResult 数据结构新增 ming_ge 和 ming_ge_desc 字段，扩展八字计算结果的数据范围

## Impact

- **后端**：`pkg/bazi/engine.go`（BaziResult 结构体 + Calculate 函数调用）、新增 `pkg/bazi/mingge.go`
- **前端**：`pages/ResultPage.tsx`（顶部 Badge 渲染 + Modal 弹出逻辑）、`pages/ResultPage.css`（格名 Badge 样式）
- **API**：`POST /api/bazi/calculate` 返回值新增 ming_ge / ming_ge_desc 字段（向后兼容，旧命盘快照 result_json 中无此字段则前端降级不展示）
- **数据库**：无需变更表结构，result_json JSONB 字段自动承接新字段
