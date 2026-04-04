## 1. 后端 Repository 层

- [x] 1.1 在 `admin_repository.go` 新增 `ListAIRequestLogs(page, pageSize int, statusFilter string)` 函数，JOIN `llm_providers` 取 Provider 名称，返回日志列表和总数
- [x] 1.2 在 `admin_repository.go` 新增 `GetAILogsSummary()` 函数，查询近 7 天按天分组的调用统计（total/success/error/avg_duration_ms），无数据日期补零

## 2. 后端 Handler 层

- [x] 2.1 在 `admin_handler.go` 新增 `AdminListAILogs` handler，解析 page/status 查询参数，调用 Repository 返回分页结果
- [x] 2.2 在 `admin_handler.go` 新增 `AdminGetAILogsSummary` handler，调用 Repository 返回统计摘要

## 3. 后端路由注册

- [x] 3.1 在 `cmd/api/main.go` 的 Admin 路由组中注册 `GET /api/admin/ai-logs` 和 `GET /api/admin/ai-logs/summary`

## 4. 前端 API 层

- [x] 4.1 在 `src/lib/adminApi.ts` 新增 `getAILogs(page, status)` 和 `getAILogsSummary()` 请求函数

## 5. 前端页面

- [x] 5.1 创建 `src/pages/admin/AdminAILogsPage.tsx`，包含：顶部统计卡片、状态筛选栏、日志明细表格（支持展开错误详情）、分页器
- [x] 5.2 在 Admin 路由中注册 `/admin/ai-logs` 路由并在侧边栏新增导航入口

## 6. 验证

- [x] 6.1 编译后端代码确保无报错
- [ ] 6.2 在线上环境触发一次 AI 调用后，验证日志页面能正确显示记录
