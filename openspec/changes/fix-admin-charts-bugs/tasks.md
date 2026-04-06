## 1. 后端：修复 SQL 重复行问题

- [x] 1.1 修改 `backend/internal/model/admin.go`，在 `AdminChartRecord` 结构体中新增 `AIResultStructured *json.RawMessage` 字段（json tag: `ai_result_structured`）
- [x] 1.2 修改 `backend/internal/repository/admin_repository.go`，将 `ListBaziCharts` 的 `LEFT JOIN ai_reports` 改为两个相关子查询：①取最新 `content` ②取最新 `content_structured`，消除重复行
- [x] 1.3 同步更新 `rows.Scan(...)` 调用，新增扫描 `AIResultStructured` 字段

## 2. 前端：修复 loading 竞态问题

- [x] 2.1 修改 `frontend/src/pages/admin/AdminChartsPage.tsx`，将 `liunianLoading: boolean` 改为 `liunianLoading: Record<string, boolean>`
- [x] 2.2 更新 `useEffect` 中设置 loading 的逻辑，改为 `setLiunianLoading(prev => ({ ...prev, [expandedId]: true/false }))`
- [x] 2.3 更新模板中引用 `liunianLoading` 的判断，改为 `liunianLoading[chart.id]`

## 3. 前端：修复 AI 报告结构化展示

- [x] 3.1 在 `ChartRecord` 接口中新增 `ai_result_structured?: { personality?: string; career?: string; romance?: string; health?: string }` 字段
- [x] 3.2 将详情面板的 AI 报告区域从渲染 `chart.ai_result`（原始 JSON 字符串）改为：优先读取 `chart.ai_result_structured` 分块渲染（性格/事业/感情/健康四章节）
- [x] 3.3 新增降级逻辑：当 `ai_result_structured` 为 null 时显示"旧格式报告，无结构化内容"提示

## 4. 验证

- [ ] 4.1 本地重启后端，访问 `/api/admin/charts`，确认同一命盘多条 ai_reports 时只返回一行
- [ ] 4.2 管理后台页面展开命盘详情，确认 AI 报告显示为结构化四章节而非 JSON 文本
- [ ] 4.3 快速切换展开多个命盘，确认流年 loading 各自独立显示
