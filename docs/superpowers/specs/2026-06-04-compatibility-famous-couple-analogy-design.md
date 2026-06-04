# 合盘「名人配对类比」设计文档

日期：2026-06-04
状态：已通过 brainstorming 评审，待用户复核

## 1. 目标

在合盘结果里，用一对名人/经典情侣类比「你们这对」的关系气质，作为顶部醒目的情感钩子。
类比要贴合实际命盘动态（可甜可虐），与全站「克制直断」的口吻一致——不强行「天作之合」。

用户原话：「我的八字怎么样、对方的八字怎么样，就好像梁山伯遇上了祝英台之类的。」
注意：梁祝是悲剧 CP，所以类比要捕捉关系的**真实动态**（一见倾心 / 现实阻隔 / 单向 / 反复拉扯…），好坏都可以，而非一味奉承。

## 2. 关键决策（brainstorming 已确认）

| 决策点 | 结论 |
|---|---|
| 用在哪里 | 结果页（不是输入页的试用样例） |
| 谁来选 CP | LLM 自由发挥，不维护策展白名单 |
| 放哪、多醒目 | 顶部醒目类比卡，始终可见 |
| 生成时机 | 随「生成深度解读」LLM 调用一并产出，**不动**瞬时评分的「开始合盘」路径（零额外成本） |
| 类比粒度 | 只做 couple 级（一对 CP），**不做**「你像 X / 对方像 Y」逐人映射 |
| 分享图 / PDF | 第一版**不包含**，结果页跑通后再评估 |

## 3. 数据结构

在合盘结构化报告（`content_structured`）JSON schema 里新增一个字段。该字段随报告 JSON 自然持久化，**无需数据库迁移**。

```jsonc
"famous_couple": {
  "couple": "梁山伯与祝英台",              // 这对 CP 的名字
  "tagline": "一见倾心，却被现实层层阻隔",   // 一句话点出关系气质
  "reason": "你们的吸引力来得快而强（年支六合），但长期更受现实安排牵制——像梁祝，情分真，却容易卡在外部阻力上。"
  // 1–2 句，大白话，扣住报告里已有的具体信号，条件语气、不下绝对命运断语
}
```

- 字段全部可选/可缺省：旧报告没有该字段时前端优雅降级（见 §5）。
- 只产出一对 CP。

### 类型同步点
- Go：`backend/internal/model/compatibility.go:203` 的 `CompatibilityStructuredReport` 结构体新增 `FamousCouple` 字段（`json:"famous_couple,omitempty"`），并新增对应子结构体（`Couple` / `Tagline` / `Reason`）。
  - **为什么必须改 Go 结构体（载重项，非可选）**：`GenerateCompatibilityReport`（`compatibility_service.go:358-371`）先把 LLM 返回的 JSON `Unmarshal` 进该结构体，再 `Marshal` 回去持久化。**不在结构体里的字段会在这次 round-trip 中被静默丢弃**——所以光改 prompt 不够，不加结构体字段类比会存不进去。
- TS：`frontend/src/lib/api.ts` 的 `CompatibilityStructuredReport` 接口新增可选 `famous_couple` 字段及其类型。

## 4. Prompt 约束（写入 `backend/pkg/prompt/canonical_compatibility.go`）

在输出 JSON schema 里加入 `famous_couple`，并补充约束段落：

- 类比必须反映**这对关系的真实动态**（综合分、四模块分、缘分时长、性格差异 fit/clash），而非一律浪漫圆满；数据偏负向时可以是苦情/悲剧 CP。
- 选广为人知的情侣（真实或传说皆可，LLM 自由联想），**得体、不越线、不出现不合适或冒犯性的配对**。
- `couple` 给名字，`tagline` 一句话气质，`reason` 1–2 句大白话、落到「你们 / 相处」、引用报告里已有的信号、用条件语气。
- 首次出现的八字术语仍遵循全局「术语必须紧跟大白话解释」的约束。

## 5. 前端：顶部类比卡

新增组件 `FamousCoupleCard`（含配套 CSS），挂载在 `CompatibilityResultPage` 里
`CompatibilityStickyHeader` 之后、`SectionVerdict` 附近的顶部位置。

数据源：`detail.latest_report?.content_structured?.famous_couple`。
生成动作复用现有 `handleGenerateReport`（与 `DeepReportNarrative` 同一入口）。

三种状态：

| 状态 | 判定 | 显示 |
|---|---|---|
| 还没生成深度解读 | `!detail.latest_report` | 钩子占位：「✨ 生成深度解读，揭晓你们的名人配对」+ 生成按钮（点了走 `handleGenerateReport`，loading 时禁用） |
| 已生成且有类比 | `famous_couple` 存在 | 醒目类比卡：大字 `couple` + `tagline` + 一行 `reason` |
| 旧报告但无该字段 | `latest_report` 存在但 `famous_couple` 缺省 | 隐藏该卡（不误导、不重复提示生成） |

## 6. 改动清单

- `backend/pkg/prompt/canonical_compatibility.go`：schema 加 `famous_couple` + 约束段落。
- `backend/internal/model/compatibility.go`：`CompatibilityStructuredReport` 加 `FamousCouple` 字段 + 子结构体（**必须**，否则 round-trip 丢字段，见 §3）。
- `frontend/src/lib/api.ts`：`CompatibilityStructuredReport` 加可选 `famous_couple` 类型。
- 新增 `frontend/src/components/compatibility/FamousCoupleCard.tsx` + `.css`。
- `frontend/src/pages/CompatibilityResultPage.tsx`：在顶部挂载 `FamousCoupleCard`，传入 `structuredReport`、`hasReport`、`reportLoading`、`onGenerateReport`。

**不改动**：合盘评分引擎、create reading 服务路径、数据库 schema/迁移、分享图 / PDF 导出。

## 7. 成功标准（可验证）

1. 生成深度解读后，结果页顶部出现名人类比卡，显示 CP 名 + 气质 + 一句理由，理由扣住该盘实际信号。
2. 未生成深度解读时，顶部显示生成钩子，点击可触发生成。
3. 旧报告（无该字段）不显示空卡、不误导。
4. 刷新页面后类比卡内容保持不变（随报告持久化）。
5. 负向/低分关系能给出苦情或谨慎类比，而非强行圆满。

## 8. 范围外 / 后续可选

- 分享图片 / PDF 导出中包含名人类比。
- 逐人映射（你像 X / 对方像 Y）。
- 输入页的「一键试用名人样例八字」。
