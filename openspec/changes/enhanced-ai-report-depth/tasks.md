## 1. 数据库 Migration

- [x] 1.1 在 `backend/pkg/database/database.go` 的 Migrate 函数中新增 `ALTER TABLE ai_reports ADD COLUMN IF NOT EXISTS content_structured JSONB`
- [x] 1.2 在 `backend/internal/model/model.go` 的 `AIReport` 结构体中新增 `ContentStructured *json.RawMessage \`json:"content_structured,omitempty"\`` 字段

## 2. 重写 Prompt 构建函数

- [x] 2.1 重写 `backend/internal/service/report_service.go` 的 `buildBaziPrompt()`，新增「===第一步：命局分析总览===」章节（要求 AI 输出推理逻辑，术语+白话双层，不限制为「心中完成」）
- [x] 2.2 修改 Prompt 中五章节要求：每章要求同时提供 `brief`（约100字摘要）和 `detail`（约350字含推理依据，术语+白话并行）
- [x] 2.3 更新第三步输出格式约束，要求严格输出新 JSON 结构：`{ yongshen, jishen, analysis: {logic, summary}, chapters: [{title, brief, detail}×5] }`
- [x] 2.4 将 `callOpenAICompatible` 中 `MaxTokens` 从 3500 提升到 4500，以容纳更长的结构化输出

## 3. 重写 JSON 解析与存储逻辑

- [x] 3.1 在 `GenerateAIReport()` 中定义新结构体 `structuredReport`（含 `Analysis` + `Chapters` 字段）
- [x] 3.2 重构 JSON 解析逻辑：优先使用新结构体解析，成功后将 chapters[].brief 拼接为纯文字写入 `content` 字段（兜底），完整 JSON 写入 `content_structured`
- [x] 3.3 保留旧三层兜底解析链（Markdown 剥离 → 正则提取 → 原始文本），但降级仅写 `content`，`content_structured` 置 nil
- [x] 3.4 更新 `repository.CreateReport()` 签名，接收 `contentStructured *json.RawMessage` 参数，写入新字段

## 4. 更新 API 响应

- [x] 4.1 确认 `AIReport` Model 新字段在 `GET /api/bazi/history/:id` 历史记录详情响应中正确序列化（含 `content_structured`）
- [x] 4.2 确认 `POST /api/bazi/report` 的响应中 `report` 对象包含 `content_structured` 字段

## 5. 前端报告区域重构

- [x] 5.1 在 `frontend/src/lib/api.ts` 中更新 `AIReport` TypeScript 接口，新增 `content_structured` 可选字段（定义对应 TS 类型 `StructuredReport`）
- [x] 5.2 在 `ResultPage.tsx` 报告区域顶部新增「精简 / 专业」切换按钮（`useState` 管理 `mode: 'brief' | 'detail'`，默认 `'brief'`）
- [x] 5.3 实现精简模式渲染：展示 `analysis.summary` + 各章 `brief` 列表
- [x] 5.4 实现专业模式渲染：展示 `analysis.logic`（带标题「命局分析总览」）+ 各章 `detail` 列表
- [x] 5.5 实现降级渲染：`content_structured` 为空时隐藏切换按钮，直接渲染 `content` 纯文字（保持现有逻辑）
- [x] 5.6 为切换按钮和两种模式添加 CSS 样式（使用 CSS Variables，放入 `ResultPage.css`）

## 6. 验证

- [x] 6.1 本地触发生成新报告，确认 DB 中 `content_structured` 字段有效写入
- [x] 6.2 在前端验证精简/专业切换 UI 正常工作
- [x] 6.3 读取一条历史旧报告（`content_structured` 为 NULL），验证降级渲染路径无报错
- [x] 6.4 模拟 AI 返回非 JSON 格式，验证兜底路径正确写入 `content` 且 `content_structured` 为 NULL
