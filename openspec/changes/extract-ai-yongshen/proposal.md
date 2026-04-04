## Why

目前八字引擎（`bazi_engine.go`）在重构后去除了硬编码的喜用神和忌神计算逻辑，转而通过 LLM 在生成解读报告时进行推断。然而，由于 LLM 仅输出纯文本格式的报告内容，后端并未对其进行解析提取，导致前端的“喜用”和“忌神”徽章长期显示为空白。为了解决数据断层问题，我们需要从 AI 的输出中结构化地提取出喜用神/忌神并存储回数据库中。

## What Changes

1. 修改 `report_service.go` 中提供给生成八字报告大模型的 Prompt，明确要求其使用严格的 JSON 结构输出（包含 `yongshen`, `jishen`, 和 `report` 字段）。
2. 在 `GenerateAIReport` 服务层接收大模型的文本响应后，剥离并解析可能存在的 markdown 代码块，进而解析出 JSON 对象。
3. 若解析成功，服务层将调用数据访问层把 `yongshen` 和 `jishen` 更新到 `bazi_charts` 数据表。
4. 前端需要针对空态情况处理，比如当 `yongshen` 尚未生成时，不再显示毫无内容的带颜色徽章。

## Capabilities

### Modified Capabilities
- `bazi-engine`: 需要在 AI 生成侧增强带有结构化元数据的输出能力。

## Impact

- `backend/internal/service/report_service.go`: Prompt 更新及 JSON 提取逻辑
- `backend/internal/repository/chart_repository.go` (或其他保存逻辑): 新增/更新图表的 `Yongshen` 和 `Jishen`
- `frontend/src/pages/ResultPage.tsx`: 对于尚未获得 `yongshen` 和 `jishen` 内容时的样式平滑处理（可能在生成期间展示 Loading 或隐藏）
