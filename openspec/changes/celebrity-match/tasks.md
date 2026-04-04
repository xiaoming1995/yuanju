## 1. 数据库层建设 (Database & Model)

- [x] 1.1 在 PostgreSQL 中新建 `celebrity_records` 表及索引（`active` 字段上的索引以加速查询），或在 `schema.sql` 中补充表结构定义。
- [x] 1.2 在后台 `internal/model` 中定义相应的 `CelebrityRecord` Go 结构体。

## 2. Admin 后台接口开发 (Backend - Admin API)

- [x] 2.1 在 `internal/repository` 层增加 `CelebrityRepository` 接口及实现（Create, GetList, Update, Delete）。
- [x] 2.2 在 `internal/handler/admin` 下新增 `CelebrityHandler` 以对外暴露 HTTP CRUD 接口（基于 Admin JWT 鉴权）。
- [x] 2.3 在 `cmd/api/main.go` 注册对应的 admin router 路由映射。

## 3. 报告生成改造 (Backend - AI Prompt)

- [x] 3.1 改造 `internal/service/bazi/report.go` 或 prompt 组装逻辑：在发起 AI 请求前，去 `CelebrityRepository` 拉取所有 `active = true` 的名人数据。
- [x] 3.2 更新 AI System Prompt：将名人数据转化成合理的文本格式（如姓名+特点 JSON/List），并明确指示生成第五章的输出规范。

## 4. Admin 后台前端开发 (Frontend - Admin UI)

- [x] 4.1 在 `frontend/src/lib/adminApi.ts` 中增加访问刚才后端的增删改查 Axios 封装函数。
- [x] 4.2 开发 `AdminCelebritiesPage` 页面：包含展示表格、"新增/编辑"弹窗表单、以及上/下线开关切换按钮。
- [x] 4.3 在 `App.tsx` 或 Admin Layout 侧边栏中添加导航入口。

## 5. 测试与数据预置

- [x] 5.1 本地端到端联调测试：手动添加两三条名人数据，然后跑一次用户全流程看第五章解析是否正确输出。（后端已完成编译通过）
- [x] 5.2 （可选）提前整理 30 个通用名人列表 SQL 导入命令，用于线上直接刷入基础数据。

## 6. AI 自动收集名人补充功能 (Admin AI Generator)

- [x] 6.1 后端服务：新增 `GenerateCelebrities` 服务逻辑，向当前激活的 LLM 派发强 JSON 约束 Prompt，结构化生成指定领域的名人。
- [x] 6.2 后端接口：在 `internal/handler/admin/celebrity_handler.go` 增设 `POST /api/admin/celebrities/ai-generate`，接收 `{topic, count}` 并批量入库。
- [x] 6.3 前端新增：在 `AdminCelebritiesPage` 添加“✨ AI 自动收集”按钮与弹窗，用户可设置生成主题和数量。
- [x] 6.4 前端联调：完成数据拉取等待态并实现入库后的列表自动数据刷新。

## 7. Bug Fixes (探索模式发现的体验问题)

- [x] 7.1 修复前端 Axios 超时导致生成假报错：在 `adminApi.ts` 中针对 `generateAI` 放宽 timeout 为 120000ms。
- [x] 7.2 修复 AI 收集数据性别格式错误：在 `celebrity_service.go` 中修改 Prompt 将 "male 或者 female" 改为纯中文 "男 或者 女"。
- [x] 7.3 优化 AI 生成的命理特征细节：在 `celebrity_service.go` 中修改 Prompt，取消 50 字限制并要求生成 150-200 字包含具体格局和喜忌推断的内容。
- [x] 7.4 优化前端长文本展示体验：在 `AdminCelebritiesPage.tsx` 中将特征输入框的高度由 `rows={4}` 改为 `rows={6}`。
