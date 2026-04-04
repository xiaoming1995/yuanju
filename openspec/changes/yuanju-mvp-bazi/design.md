## Context

缘聚是一个全新项目，从零开始构建。技术栈沿用团队熟悉的方案：Go 后端 + React 前端。MVP 聚焦八字功能，核心挑战在于：

1. **八字算法的准确性**：阳历→干支四柱的转换涉及复杂的历法计算
2. **AI 报告的一致性**：大模型输出需要结构化和风格统一
3. **双层 UI 的体验平衡**：普通用户与专业用户的信息密度需要分层设计

## Goals / Non-Goals

**Goals:**
- 实现完整的八字四柱计算链路（阳历 → 农历 → 天干地支 → 五行分析）
- 提供 AI 自然语言八字解读，输出通俗、准确的个性化报告
- 支持用户注册登录、历史记录保存
- 前端双层视图：简洁解读视图 + 专业命盘视图
- Docker 容器化，支持快速部署

**Non-Goals:**
- 紫微斗数、西洋占星（MVP 后阶段）
- 合盘配对、命理师入驻（MVP 后阶段）
- 付费订阅与商业化（MVP 后阶段）
- 移动端 App（Web 响应式即可）

## Decisions

### D1：八字算法实现方式

**决策**：自研 Go 实现八字算法核心，不依赖第三方命理 API。

**理由**：
- 第三方 API 存在成本与可控性风险
- 八字四柱计算逻辑相对固定，核心是历法转换 + 天干地支映射
- 开源社区有成熟的农历算法（寿星万年历）可参考

**方案对比**：
| 方案 | 准确性 | 速度 | 依赖风险 |
|------|--------|------|---------|
| 自研算法 | 高 | 快 | 低 |
| 第三方 API | 高 | 中 | 高 |
| 纯 AI 生成 | 中 | 慢 | 中 |

---

### D2：AI 报告生成策略

**决策**：算法计算出结构化八字数据，再通过 Prompt 工程交由大模型生成自然语言报告。

**流程**：
```
用户输入生辰 → Go 算法引擎计算 → 结构化八字数据
                                        ↓
                               构造富含上下文的 Prompt
                                        ↓
                               大模型 API（DeepSeek 优先）
                                        ↓
                               自然语言报告（分段：性格/感情/事业/健康）
```

**理由**：算法保证数据准确，AI 保证表达通俗。二者分工明确，互不干扰。

---

### D3：大模型 API 选型

**决策**：主力使用 DeepSeek API，OpenAI 作为备选。

**理由**：DeepSeek 中文理解能力强，成本显著低于 OpenAI，且命理类文本生成效果优秀。

---

### D4：数据库设计

**核心表结构**：
```sql
users          -- 用户信息（id, email, password_hash, created_at）
bazi_charts    -- 八字命盘（id, user_id, birth_datetime, four_pillars JSON, wuxing JSON）
ai_reports     -- AI解读报告（id, chart_id, content, model, created_at）
```

**理由**：将命盘与 AI 报告分表存储，支持未来对同一命盘生成多版本报告，也便于缓存控制。

---

### D5：前端架构

**决策**：使用 React + Vite，CSS 变量实现设计系统，不引入 UI 框架。

**双层视图实现**：
```
BaziPage
├── QuickSummaryCard     ← 普通用户：AI解读摘要（默认展开）
├── ChartSection         ← 专业视图：四柱命盘图（默认折叠，可展开）
│   ├── FourPillarsGrid
│   ├── WuxingRadar      ← 五行雷达图
│   └── DayunTimeline    ← 大运时间轴
└── FullReportSection    ← 完整 AI 报告（性格/感情/事业/健康分段）
```

---

### D6：API 设计

```
POST /api/auth/register
POST /api/auth/login
GET  /api/auth/me

POST /api/bazi/calculate      ← 计算八字（无需登录可用）
POST /api/bazi/report         ← 生成 AI 报告（需登录）
GET  /api/bazi/history        ← 历史记录（需登录）
GET  /api/bazi/history/:id    ← 历史详情
```

## Risks / Trade-offs

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| 八字算法存在边界 case 错误 | 中 | 高 | 引入知名八字案例作为单元测试基准 |
| AI 报告质量不稳定 | 中 | 中 | Prompt 模板固定结构，对输出做长度和格式校验 |
| DeepSeek API 限流/故障 | 低 | 中 | 接入 OpenAI 作为 fallback，报告生成异步化 |
| 用户对 AI 解读产生信任疑虑 | 中 | 低 | 报告页面注明"AI 辅助解读，仅供参考" |

## Open Questions

- 八字时辰的早子时/晚子时处理规则需要确认（影响日柱计算）
- AI 报告是否需要流式输出（SSE）以改善等待体验？MVP 阶段可先做同步，后续优化
- 历史记录是否支持分享给他人？MVP 暂不做
