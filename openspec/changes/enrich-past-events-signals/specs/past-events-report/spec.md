## MODIFIED Requirements

### Requirement: Admin Prompt 管理支持 past_events 模块
系统 SHALL 在 prompt 管理体系中支持 module=`past_events` 的 Prompt 模板，Admin 可通过现有 `PUT /api/admin/prompts/past_events` 接口修改。

Prompt 模板需接收以下模板变量：
- `{{.Gender}}`：命主性别
- `{{.DayGan}}`：日干
- `{{.NatalSummary}}`：原局四柱及十神概要
- `{{.YearsData}}`：JSON 格式的年份信号列表（每条信号含 polarity/source 字段）
- `{{.DayunList}}`：大运列表文字描述（既有）
- `{{.YongshenInfo}}`：命主用神/忌神描述（如"用神：火 / 忌神：水"）
- `{{.GejuSummary}}`：原局格局描述（如"正官格、官印相生"，无格局信息时为空串）
- `{{.DayunHuahe}}`：大运合化标签（如"第3步大运 丁壬合化木"，无则为空串）
- `{{.StrengthDetail}}`：加权身强弱评分明细（如"中和(评分2): 月令同气、地支藏比劫透出"）

#### Scenario: Prompt 模板渲染失败
- **WHEN** Prompt 模板语法错误或变量缺失导致渲染失败
- **THEN** 接口返回 500，错误信息说明"Prompt模板渲染失败"

#### Scenario: 数据库中不存在 past_events Prompt
- **WHEN** 系统启动时 `past_events` Prompt 不存在于数据库
- **THEN** `pkg/database/database.go` 启动 seed 阶段写入默认 Prompt（含全部新增模板变量）

#### Scenario: 旧版 Prompt 自动升级
- **WHEN** 数据库已有 `past_events` Prompt 但内容不含 `{{.YongshenInfo}}` 字段
- **THEN** 启动 seed 阶段自动以默认新版 Prompt 覆盖（参考 `dayun_summaries` 升级模式）

#### Scenario: 已自定义 Prompt 不被覆盖
- **WHEN** 数据库 `past_events` Prompt 已包含 `{{.YongshenInfo}}` 字段（说明已是新版或 admin 已自行加入）
- **THEN** 启动 seed 阶段不覆盖现有 Prompt

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
      "signals": ["婚恋_合", "事业"],
      "narrative": "庚寅年，偏财星庚透干为用神，男命感情星力量显现；寅木与日支相合，夫妻宫被激活。该年感情进展顺利，事业亦有贵人扶持。"
    }
  ],
  "dayun_summaries": [
    {
      "gan_zhi": "甲子",
      "themes": ["事业↑", "贵人扶持"],
      "summary": "..."
    }
  ]
}
```

`signals` 数组项可以是简单字符串（兼容旧版），也可以是结构化对象 `{"type": "...", "polarity": "..."}`（新版可选）。`narrative` 必须综合考虑 polarity 与基底色，避免吉凶颠倒。

#### Scenario: 每年叙述不超过 3 句
- **WHEN** AI 生成报告
- **THEN** 每个年份的 narrative 字段包含 2-3 句中文描述

#### Scenario: 无信号年份也有叙述
- **WHEN** 某流年算法未检测到任何信号
- **THEN** 该年的 narrative 不为空，AI 给出"该年较为平稳"类型的简短描述

#### Scenario: 用神基底与单条信号 polarity 冲突时的叙述策略
- **WHEN** 流年用神基底色为凶 但 个别信号本身为吉（如财星透干但财为忌神）
- **THEN** narrative 应以基底色为主，注明转折语（如"虽现财星，然为忌神，反主破耗"），避免直白写"财运提升"

#### Scenario: 大运 summary 引用合化与神煞
- **WHEN** 当前大运被算法标注为合化大运 或 含强神煞
- **THEN** 该大运 summary 的 themes 与 summary 文字必须包含合化方向与神煞影响（如"丁壬合化木——日主性向转向文教/创意"）

---

## ADDED Requirements

### Requirement: 用神/忌神信息注入 Prompt
系统 SHALL 在 `GeneratePastEventsStream` 中读取命盘的用神/忌神信息（来源：`bazi.Calculate` 返回的 `Yongshen` / `Jishen` 字段，缺失时降级为 `Tiaohou.Expected` 描述），构造 `YongshenInfo` 字符串注入 Prompt。

#### Scenario: 用神字段完整
- **WHEN** `BaziResult.Yongshen` 非空且 `BaziResult.Jishen` 非空
- **THEN** YongshenInfo 形如"用神：火、土 / 忌神：水、木"

#### Scenario: 仅有调候喜神
- **WHEN** `BaziResult.Yongshen` 为空但 `Tiaohou.Expected` 非空
- **THEN** YongshenInfo 形如"调候喜神：丙、丁（综合用神信息缺失，以调候为参考）"

#### Scenario: 用神信息全部缺失
- **WHEN** 上述两类信息均空
- **THEN** YongshenInfo 为空串，Prompt 不强制 AI 引用

---

### Requirement: 大运合化与身强弱明细注入 Prompt
系统 SHALL 在 `GeneratePastEventsStream` 中根据信号引擎输出的大运合化结果与加权身强弱评分，分别构造 `DayunHuahe` 与 `StrengthDetail` 字符串注入 Prompt。

#### Scenario: 当前命盘存在大运合化
- **WHEN** 任一大运被信号引擎判定为"合而成化"
- **THEN** DayunHuahe 形如"第3步大运 丁壬合化木：本段大运日主性向偏木"

#### Scenario: 多步大运均存在合化
- **WHEN** 多步大运均被判定合化
- **THEN** DayunHuahe 多行换行拼接，每行一条

#### Scenario: 加权身强弱评分输出
- **WHEN** 信号引擎完成加权身强弱评分
- **THEN** StrengthDetail 形如"中和(评分2): 月令同气加分5，地支克泄减分3"

#### Scenario: 字段为空时不污染 Prompt
- **WHEN** 无合化或评分缺失
- **THEN** DayunHuahe 或 StrengthDetail 为空串，模板渲染时不输出多余空白段落
