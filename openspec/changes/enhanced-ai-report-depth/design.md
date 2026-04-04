## Context

缘聚平台使用 `report_service.go` 中的 `buildBaziPrompt()` 构建 Prompt，调用 LLM 生成八字命理报告。当前 Prompt 将 AI 推理过程（CoT）设为「心中完成、不输出」，报告输出为纯文字字符串存入 `ai_reports.content TEXT` 字段。探索阶段明确了新需求：推理链条可见、精简/专业双模式、结构化 JSON 存储。

## Goals / Non-Goals

**Goals:**
- 重写 Prompt，让 AI 输出完整结构化 JSON：含命局推理总览（`analysis`）+ 每章精简+详细双版本（`chapters[]`）
- 新增 `ai_reports.content_structured JSONB` 字段，存储结构化报告，旧 `content` 字段保留做兜底
- 前端报告区域新增「精简 / 专业」模式切换，前端读取不同字段渲染
- 推理语言采用「术语 + 白话」双层：先写专业推导，紧跟一句通俗解释

**Non-Goals:**
- 不做 Streaming 流式输出（单独变更处理）
- 不做多轮追问对话
- 不迁移历史 `content` 数据到新格式（历史报告只走旧渲染路径）

## Decisions

### D1：新 JSON 输出结构

```json
{
  "yongshen": "水金",
  "jishen": "木火",
  "analysis": {
    "logic": "命局推理总览（600字内，术语+白话双层）",
    "summary": "一句话命局核心特质"
  },
  "chapters": [
    {
      "title": "性格特质",
      "brief": "精简版约100字",
      "detail": "详细版约350字，含推理依据（术语+白话）"
    },
    { "title": "感情运势", "brief": "...", "detail": "..." },
    { "title": "事业财运", "brief": "...", "detail": "..." },
    { "title": "健康提示", "brief": "...", "detail": "..." },
    { "title": "大运走势", "brief": "...", "detail": "..." }
  ]
}
```

旧 Prompt 输出 `{"yongshen","jishen","report"}` 三字段，新 Prompt 输出六字段。`report` 字段废弃，兜底时将 `chapters[].brief` 拼接为纯文字写入 `content`。

**决策理由**：前端精简/专业模式各取 `brief`/`detail`，无需二次 AI 调用，一次生成全包含。

### D2：数据库新增 JSONB 字段（方案 Y）

```sql
ALTER TABLE ai_reports ADD COLUMN content_structured JSONB;
```

- 旧 `content TEXT` 保留，读取时优先使用 `content_structured`，无则降级到 `content`
- 新报告同时写入两个字段：`content_structured` 写完整 JSON，`content` 写所有 `brief` 拼接的纯文字（兜底）

**决策理由**：不破坏历史数据结构，表结构清晰，降级逻辑简单。

### D3：JSON 解析策略保持三层兜底

现有三层解析（直接解析 → 剥离 Markdown → 正则提取）继续保留，新增对 `analysis` 和 `chapters` 字段的提取逻辑。若结构化解析失败，降级为原始文本写入 `content`，`content_structured` 置 NULL。

### D4：前端精简/专业切换

- 默认展示「精简」模式
- 切换按钮放在报告区域标题右侧
- 精简：展示 `analysis.summary` + 各章 `brief`
- 专业：展示 `analysis.logic` + 各章 `detail`
- 如果 `content_structured` 为空（旧报告 / 解析失败），直接展示 `content`，不显示切换按钮

## Risks / Trade-offs

- **Token 消耗增加**：新结构化 JSON 比旧纯文字约多 50% Token（brief+detail 双版本）→ 接受，信息密度提升值得
- **JSON 解析失败率可能略升**：结构更复杂，LLM 可能在 `chapters` 数组格式上出错 → 兜底链已覆盖
- **Prompt 长度增加**：新格式要求更详细，输入 Prompt 约增加 200 字 → MaxTokens 从 3500 提升到 4500
- **旧报告无切换按钮**：前端需做 `content_structured` 是否存在的判断 → 设计已明确

## Migration Plan

1. 执行 DDL：`ALTER TABLE ai_reports ADD COLUMN content_structured JSONB`（在 `database.go` 的 Migrate 函数中新增）
2. 部署新后端：新报告走新逻辑，旧报告不受影响
3. 前端发布：`content_structured` 空值时降级渲染，兼容旧报告

## Open Questions

- 无，探索阶段所有决策已明确
