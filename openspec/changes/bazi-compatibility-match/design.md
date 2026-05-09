## Context

当前 Yuanju 的核心链路围绕单人命盘展开：`bazi.Calculate()` 负责起盘，`bazi_charts` 保存用户命盘，`ai_reports` / `ai_liunian_reports` / `ai_past_events` 等围绕单个 `chart_id` 做扩展。这个结构适合回答“命主自己如何”，但不适合表达“双人关系如何”，因为合盘天然需要两份命盘、双盘互动信号、单独的结果生命周期，以及与普通历史记录隔离的存储模型。

同时，当前产品已经形成了稳定的方法论：

- 算法先产出结构化数据
- AI 再负责组织语言
- 登录用户才能沉淀报告历史

合盘功能应延续这条主线，而不是退化为纯 Prompt 拼接的黑箱推理。

## Goals / Non-Goals

**Goals:**
- 为婚恋场景新增一条独立的合盘资源链路，能承载双方出生信息、双方命盘快照、结构化匹配结果和 AI 报告
- 用结构化信号表达“为什么合/不合”，而不是只给单一分数
- 产出稳定的四维结果：吸引力、稳定度、沟通协同、现实磨合
- 支持登录用户查看和复用合盘历史，且不污染现有单人命盘历史
- 尽量复用既有的 `pkg/bazi` 起盘能力、AI provider、Prompt 管理与前端展示模式

**Non-Goals:**
- 不在第一版实现择日、结婚年份预测、长期大运同步推演
- 不扩展为朋友、亲子、同事等泛人际关系匹配
- 不支持时辰缺失的模糊合盘；第一版要求双方输入完整时辰
- 不强求建立一个“绝对正确”的命理打分体系；MVP 以结构化证据 + AI 解读为主

## Decisions

### D1：合盘使用独立资源模型，不复用 `bazi_charts` / `ai_reports`

**选择**：新增一组合盘专用表，而不是把“对象 B”塞进 `bazi_charts`。

建议结构：

- `compatibility_readings`
  - `id`
  - `user_id`
  - `overall_level`
  - `dimension_scores` JSONB
  - `summary_tags` JSONB
  - `analysis_version`
  - `created_at`
  - `updated_at`
- `compatibility_participants`
  - `id`
  - `reading_id`
  - `role` (`self` / `partner`)
  - `display_name`（可选，如“我”“对方”）
  - `birth_profile` JSONB
  - `chart_hash`
  - `chart_snapshot` JSONB
  - `created_at`
- `compatibility_evidences`
  - `id`
  - `reading_id`
  - `dimension`
  - `type`
  - `polarity`
  - `source`
  - `title`
  - `detail`
  - `weight`
- `ai_compatibility_reports`
  - `id`
  - `reading_id`
  - `content`
  - `content_structured` JSONB
  - `model`
  - `created_at`

**Why**：
- 防止“对象 B”进入普通 `GET /api/bazi/history`
- 合盘记录天然是“一对多子资源”（双方 + 证据 + 报告），不适合硬挂在单个 `chart_id` 上
- 未来增加“同一人和不同对象比较”时，这种模型更自然

**放弃**：
- 直接复用 `bazi_charts` 保存双方命盘：会污染单人历史、权限语义模糊
- 只保存最终报告文本：后续无法解释证据，也不利于缓存与前端细粒度展示

### D2：合盘链路沿用“结构化分析先行，AI 解读后置”

**选择**：先生成结构化合盘分析，再单独生成 AI 报告。

链路：

```text
输入双方出生信息
  -> 各自调用 bazi.Calculate()
  -> 生成双方 chart_snapshot
  -> compatibility signal engine 输出维度分数 + evidences
  -> 保存 compatibility_reading
  -> AI 基于结构化结果生成 narrative/report
```

**Why**：
- 与现有“算法精算 + AI 自然语言解读”主线一致
- 可以在不依赖 AI 的情况下先渲染结果页骨架与证据卡片
- 证据层可测试，AI 层只负责表达

**放弃**：把两份命盘直接交给大模型自由发挥。这样虽然快，但不可解释，也难以持续调优。

### D3：第一版固定四个核心维度 + 一个总体等级

**选择**：统一输出：

- `overall_level`: `high` / `medium` / `low`
- `dimension_scores`:
  - `attraction`
  - `stability`
  - `communication`
  - `practicality`

每个维度使用 `0-100` 数值，便于前端图形化渲染；总体等级由维度与关键负向证据综合得出，不单纯按平均分计算。

**Why**：
- 比“合/不合”更符合用户真实决策过程
- 比章节化散文更适合在结果页首屏快速建立感知
- 后续可以自然扩展“阶段趋势”而不破坏顶层结构

**放弃**：
- 单一总分：解释力弱
- 过多维度：第一版前端过重，运营也难讲清楚

### D4：证据模型按“维度 + 极性 + 来源”建模

**选择**：每条证据至少包含：

- `dimension`
- `type`
- `polarity` (`positive` / `negative` / `mixed` / `neutral`)
- `source`
- `title`
- `detail`
- `weight`

第一版信号源覆盖：
- 日主关系
- 五行互补 / 偏枯
- 日支（夫妻宫）合冲刑害
- 财星 / 官星等配偶相关互动
- 干支合冲刑害
- 关键神煞辅助

**Why**：
- 与 `past-events` 的 `type / polarity / source` 结构相近，团队已经熟悉
- 前端可以直接渲染成“证据标签 + 说明卡片”
- 后续 AI Prompt 可以有更稳定的输入

### D5：MVP 仅支持登录用户发起完整合盘与保存历史

**选择**：完整合盘（含 AI 报告与历史）要求登录。

**Why**：
- 合盘涉及两个人的敏感出生数据，匿名缓存和权限边界更复杂
- 当前系统已有“起盘可匿名，深度报告需登录”的模式，用户心智一致
- 历史、缓存、再次查看等场景都以登录资源为核心

**放弃**：首版 guest 合盘预览。后续如要做引流，可以再补一个“匿名只看简版、不保存历史”的阉割模式。

### D6：合盘 AI Prompt 采用独立模块 `compatibility`

**选择**：在 `ai_prompts` 中新增模块 `compatibility`，不复用 `natal` / `liunian` / `past_events` 模板。

建议输出结构：

```json
{
  "summary": "总体判断",
  "dimensions": [
    { "key": "attraction", "title": "吸引力", "content": "..." },
    { "key": "stability", "title": "稳定度", "content": "..." },
    { "key": "communication", "title": "沟通协同", "content": "..." },
    { "key": "practicality", "title": "现实磨合", "content": "..." }
  ],
  "risks": ["...", "..."],
  "advice": "..."
}
```

**Why**：
- 合盘报告的语气、章节和证据引用方式与单盘报告明显不同
- Admin 可以独立调优姻缘解读模板，而不影响单人报告

## Risks / Trade-offs

- **[命理规则争议]** 不同流派对“合盘”权重分配可能不同 → 第一版将证据与分数来源显式化，便于后续调权而不是把结论写死在文案里
- **[数据模型变复杂]** 合盘比单盘多出参与者、证据、报告三层结构 → 用独立表隔离复杂度，避免污染已有资源
- **[AI 报告与结构化结果不一致]** 若 Prompt 过松，模型可能夸大或忽略关键负向信号 → 要求 Prompt 明确引用维度分数与 evidences，结构化 JSON 输出优先
- **[时辰要求提高表单流失]** 第一版要求双方完整时辰，可能降低转化 → 先保证结论质量，缺时辰降级分析留到后续版本
- **[敏感数据合规]** 合盘天然包含“第三人”信息 → MVP 限制为登录用户私有历史，不做公开分享页默认暴露

## Migration Plan

1. 增量迁移新增合盘表、索引与默认 `compatibility` Prompt。
2. 后端新增 model / repository / service / handler / route，不改现有单人接口契约。
3. 前端新增合盘输入页、结果页、历史页入口。
4. 灰度验证通过后开放菜单入口；若回滚，仅需下线新路由并保留表结构，不影响现有单盘功能。

## Open Questions

- `display_name` 是否在第一版就让用户填写（如“我”和“对方”之外支持自定义昵称）？这不影响核心架构，可在 UI 细化时决定。
- 未来是否需要“合盘简版”与“深度版”两档输出？当前设计兼容，MVP 先只做一档完整版本。
