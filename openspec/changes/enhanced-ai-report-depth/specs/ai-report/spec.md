## MODIFIED Requirements

### Requirement: AI 报告生成接口返回格式
`POST /api/bazi/report` 的响应体中 `report` 对象 SHALL 同时包含 `content`（纯文字，向下兼容）和 `content_structured`（结构化 JSON，新增可空）两个字段。

#### Scenario: 新报告接口响应包含双字段
- **WHEN** 成功生成新格式报告并调用 POST /api/bazi/report
- **THEN** 响应 report 对象 SHALL 包含 content（字符串）和 content_structured（JSON 对象或 null）

#### Scenario: 旧接口字段向下兼容
- **WHEN** 前端读取 content_structured 为 null 的报告响应
- **THEN** 前端 SHALL 降级使用 content 字段，不出现空白或崩溃
