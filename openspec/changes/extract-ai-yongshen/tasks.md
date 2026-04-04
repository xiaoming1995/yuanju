## 1. 结构化 LLM 输出
- [x] 1.1 修改 `backend/internal/service/report_service.go` 的 prompt，明确要求输出 JSON 格式（带 `yongshen`, `jishen`, `report` 字段）。
- [x] 1.2 在 `report_service.go` 内部构建防御性的 JSON 截取逻辑（去头去尾，剥离 Markdown formatting 等）。
- [x] 1.3 编写 JSON 解析 (Unmarshal) 的逻辑和针对失败情况（降级为普通文本）的 Fallback 兜底逻辑。

## 2. 数据库存储与持久化
- [x] 2.1 修改 `backend/internal/repository/repository.go`（若是独立的 chart repository 文件则对应修改），新增或调整能够指定更新 `Yongshen` 和 `Jishen` 字段的 SQL 方法：`UpdateChartYongshenJishen`。
- [x] 2.2 在 `report_service.go` 被解析出喜忌且无错误的路径下，调用该 repository 方法将喜忌写入 `bazi_charts`。

## 3. 接口层与客户端同步
- [x] 3.1 修改 `backend/internal/handler/bazi_handler.go`，在请求 `api/bazi/report` 并阻塞获得结果后，确保向前端下发的 `chart` 级联包含最新的 `Yongshen` 和 `Jishen` 数据。
- [x] 3.2 优化前端 `frontend/src/pages/ResultPage.tsx`，当未获得喜用数据（即为空 `""`）时，屏蔽默认带配色的 badge 样式或者显示“AI推算中...”，而在 AI 生成完报告更新 state 之后呈现出正确的徽章。
