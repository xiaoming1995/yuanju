## 1. 后端：扩展 Prompt 模板与 Service

- [x] 1.1 在 `internal/model/model.go` 的 `PastEventsTemplateData` 结构体中新增 `DayunList string` 字段
- [x] 1.2 在 `internal/service/report_service.go` 的 `GeneratePastEventsStream` 中，将 `result.Dayun` 序列化为 `DayunList` 文字描述（格式：`大运 甲子 12-22岁 [官星/正印]`），注入模板数据
- [x] 1.3 在 `pkg/seed/seed.go`（或通过 Admin UI）更新 `past_events` Prompt 模板：新增 `{{.DayunList}}` 占位符，并在指令中要求 AI 额外输出 `dayun_summaries` 数组（含 `gan_zhi`、`themes[]`、`summary`），提供 JSON schema 示例和长度约束（每条 summary ≤ 120 字）

## 2. 前端：解析与渲染

- [x] 2.1 在 `PastEventsPage.tsx` 中新增 `DayunSummary` 接口：`{ gan_zhi: string; themes: string[]; summary: string }`
- [x] 2.2 在 `PastEventsPage` state 中新增 `summaries: DayunSummary[]`，并在 `parseAndSet` 中解析 `parsed.dayun_summaries ?? []`
- [x] 2.3 在大运分组渲染逻辑中，根据 `dayun_gan_zhi` 从 `summaries` 中查找对应总结，若存在则在大运标题下方渲染总结块（主题标签行 + 叙述文字段）
- [x] 2.4 主题标签样式：复用 `SIGNAL_LABEL` 颜色映射（已知信号类型），未知词汇使用默认色 `var(--text-muted)`；标签与年份信号标签保持视觉一致
