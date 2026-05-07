## ADDED Requirements

### Requirement: 过往事件推算 SSE 接口
系统 SHALL 提供 `POST /api/bazi/past-events-stream/:chart_id` 接口，需用户登录（`middleware.Auth()`），用于流式生成过往年份事件推算报告。

请求无 body。响应为 `text/event-stream`，data 字段格式与现有 `report-stream` 接口一致（累积 token 字符串，最后一条 `data: [DONE]`）。

#### Scenario: 缓存命中直接返回
- **WHEN** 该 chart_id 已有缓存的过往事件报告
- **THEN** 接口直接以 SSE 方式推送缓存内容，不重新调用 AI

#### Scenario: 首次生成流式推送
- **WHEN** 该 chart_id 尚无缓存
- **THEN** 接口调用算法检测所有过往流年信号，构建 Prompt，SSE 流式返回 AI 输出，生成完毕后入库缓存

#### Scenario: 无权限访问他人命盘
- **WHEN** 请求者的 user_id 与 chart 的 user_id 不匹配
- **THEN** 返回 403

#### Scenario: chart_id 不存在
- **WHEN** 指定的 chart_id 在数据库中不存在
- **THEN** 返回 404

---

### Requirement: 过往事件报告数据结构
系统 SHALL 将生成的报告以 JSONB 格式存入 `ai_past_events` 表，`content_structured` 字段为以下结构：

```json
{
  "years": [
    {
      "year": 2010,
      "age": 20,
      "gan_zhi": "庚寅",
      "dayun_gan_zhi": "甲子",
      "signals": ["婚恋", "事业"],
      "narrative": "庚寅年，偏财星庚透干，男命感情星力量显现；寅木与日支相合，夫妻宫被激活。该年感情方面或有重要进展，宜把握缘分。"
    }
  ]
}
```

#### Scenario: 每年叙述不超过 3 句
- **WHEN** AI 生成报告
- **THEN** 每个年份的 narrative 字段包含 2-3 句中文描述

#### Scenario: 无信号年份也有叙述
- **WHEN** 某流年算法未检测到任何信号
- **THEN** 该年的 narrative 不为空，AI 给出"该年较为平稳"类型的简短描述

---

### Requirement: Admin Prompt 管理支持 past_events 模块
系统 SHALL 在 prompt 管理体系中支持 module=`past_events` 的 Prompt 模板，Admin 可通过现有 `PUT /api/admin/prompts/past_events` 接口修改。

Prompt 模板需接收以下模板变量：
- `{{.Gender}}`：命主性别
- `{{.DayGan}}`：日干
- `{{.NatalSummary}}`：原局四柱及十神概要
- `{{.YearsData}}`：JSON 格式的年份信号列表

#### Scenario: Prompt 模板渲染失败
- **WHEN** Prompt 模板语法错误或变量缺失导致渲染失败
- **THEN** 接口返回 500，错误信息说明"Prompt模板渲染失败"

#### Scenario: 数据库中不存在 past_events Prompt
- **WHEN** 系统启动时 `past_events` Prompt 不存在于数据库
- **THEN** `seed.SeedLLMProviders` 或启动 seed 阶段写入默认 Prompt
