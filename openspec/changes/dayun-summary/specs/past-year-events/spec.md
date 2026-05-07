## MODIFIED Requirements

### Requirement: AI 响应 JSON 结构
系统 SHALL 期望 AI 返回以下扩展 JSON 结构：

```json
{
  "years": [...],
  "dayun_summaries": [
    {
      "gan_zhi": "甲子",
      "themes": ["事业↑", "贵人"],
      "summary": "甲子大运，官印相生..."
    }
  ]
}
```

`dayun_summaries` 字段为可选，缺失时系统 SHALL 正常工作（向后兼容）。

#### Scenario: 扩展 JSON 解析成功
- **WHEN** AI 返回含 `dayun_summaries` 的 JSON
- **THEN** `parseAndSet` 同时解析 `years` 和 `dayun_summaries`，两者分别存入对应 state

#### Scenario: 缺少 dayun_summaries 时不报错
- **WHEN** AI 返回仅含 `years` 的旧格式 JSON
- **THEN** `dayun_summaries` 解析结果为空数组，页面正常渲染年份卡片

### Requirement: Prompt 模板变量
Prompt 模板 SHALL 支持 `{{.DayunList}}` 变量，内容为大运列表的文字描述（含干支、十神、起止年龄）。

#### Scenario: DayunList 注入
- **WHEN** `GeneratePastEventsStream` 构建 Prompt
- **THEN** 模板数据中包含 `DayunList` 字段，格式为每行一条大运信息
