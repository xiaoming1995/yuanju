# 按八字起盘（四柱反推日期）设计

> 状态：设计已确认，待写实现计划

## 背景与问题

现在起盘只有一条输入路径：用户填「年月日时 + 性别」，后端用 `lunar-go` 正推四柱、五行、用神、格局、**大运**等。

存在一种常见场景：用户手里只有**四柱八个字**（别处算好的、或只看到一张排盘图），并不知道真实出生年月日时，无法用现有表单起盘。

### 核心约束：大运必须有真实出生时刻

大运起运是「出生时刻到下一个节气的天数」换算来的（精确到时分），**四柱本身推不回这个信息**。同时：

- 五行、用神、格局、柱与柱之间的神煞 —— 只靠四柱即可计算。
- 大运 / 部分流年 —— 必须有真实出生时刻。
- 同一组四柱：年柱 60 年一轮回，且时辰内 2 小时无法定位到分钟，因此「四柱 → 出生时刻」是一对多。

`BaziResult`、入库模型、报告服务到处依赖 `BirthYear/Month/Day/Hour`。任何「绕过日期的纯八字降级通路」都要在大量下游打补丁。

## 方案：反推真实公历日期，复用现有全流程

把「只有八字」**收敛回「已知真实日期」**，从而下游一行不改。

关键洞察：年柱 60 年一轮回，所以在任意 60 年窗口内，「年月日时四柱」几乎唯一对应一个公历日期。在 1900–2030 范围搜，通常只命中 **2–3 个候选日期**（相差约 60 岁），用户按「这个人大概多少岁」即可点选。

### 数据流（两步）

```
用户在首页「按八字」面板选 4 个干支 + 性别（+ 可选大致年代）
      │
      ▼  POST /api/bazi/resolve-pillars   （新增，只读，无副作用，不入库）
后端反查 [minYear, maxYear] 内能产生这组四柱的公历日期 → 候选列表（通常 2–3 个）
      │
      ▼  前端候选选择
            · 0 个 → 提示「这组八字找不到对应的真实日期，请核对」
            · 1 个 → 自动选中，直接进入下一步
            · 多个 → 弹层列出「公历日期 + 参考年龄」，用户点选
      │
      ▼  POST /api/bazi/calculate   （现有接口，零改动）
        用「选中日期 + 时辰中点小时 + 性别」调用 → 现有结果页
```

下游大运 / 神煞 / 用神 / 格局 / 流年 / AI 报告 / 入库**全部复用现成代码**。

## 后端设计

### 新增反查函数

文件：`backend/pkg/bazi/resolve_pillars.go`（新建）

```go
// Candidate 一个能产生目标四柱的候选公历日期
type Candidate struct {
    Year, Month, Day int // 公历年月日
    Hour             int // 该时辰的中点小时（如午时=12），用于后续 Calculate
    LunarDate        string // 农历表述，供前端展示（如「乙巳年六月初九」）
    RefAge           int    // 参考年龄 = referenceYear - Year，供用户按年龄辨识
}

// ResolvePillars 反查能产生目标四柱的公历日期。
// 入参为 4 个干支字符串（如 "甲子"），minYear/maxYear 为搜索范围，
// referenceYear 用于计算候选的参考年龄。
// 返回所有命中候选，按年份升序。非法 / 不自洽的四柱返回空切片。
func ResolvePillars(yearGZ, monthGZ, dayGZ, hourGZ string,
    minYear, maxYear, referenceYear int) []Candidate
```

**算法（利用循环节缩小搜索空间，避免逐日暴力）**：

1. 校验 4 个入参都是 60 甲子中的合法组合，否则直接返回空。
2. 用日柱锁定候选日：日柱是连续 60 天循环，从范围内任一已知日柱日期推算，范围内匹配目标日柱的公历日只有约 `天数/60` 个（1900–2030 约 800 个）。
3. 对每个候选日，构造 `Solar`（小时取目标时支对应时辰中点）→ `GetLunar().GetEightChar()`，比较年/月/日/时四柱是否全等。全等则计入候选（跨柱自洽性在此一并校验——月干与年干、时干与日干不符的输入不会有任何命中）。
4. 时支 → 时辰中点小时：`中点小时 = 地支序号 × 2`（子=0、丑=2、寅=4 … 亥=22）。子时统一按晚子时（0 时）处理。
5. 填 `LunarDate`（用 lunar-go 取农历串）与 `RefAge = referenceYear - Year`，按年份升序返回。

> 性能：约 800 次 `EightChar` 比对，单次请求亚秒级，无需缓存。

> `referenceYear`：由 handler 传入当前年份（从请求时间取），保持 `pkg/bazi` 纯函数、不直接读时钟。

### 新增 handler 与路由

文件：`backend/internal/handler/bazi_handler.go`（在现有 `Calculate` 旁新增）

```go
type ResolvePillarsInput struct {
    YearPillar  string `json:"year_pillar" binding:"required"`
    MonthPillar string `json:"month_pillar" binding:"required"`
    DayPillar   string `json:"day_pillar" binding:"required"`
    HourPillar  string `json:"hour_pillar" binding:"required"`
    MinYear     int    `json:"min_year"` // 可选，缺省 1900
    MaxYear     int    `json:"max_year"` // 可选，缺省 2030
}

// ResolvePillars 四柱反查候选公历日期（无需登录，只读）
func ResolvePillars(c *gin.Context)
```

- 缺省范围：`MinYear=1900`、`MaxYear=2030`（定义为包内常量）。
- `referenceYear` 取服务器当前年份传入 `bazi.ResolvePillars`。
- 返回 `{ "candidates": [...] }`；空数组是正常结果（前端提示无匹配），非错误。

路由（`backend/cmd/api/main.go`，与现有 `bazi.POST("/liu-yue", ...)` 同样无中间件）：

```go
bazi.POST("/resolve-pillars", handler.ResolvePillars)
```

### 不改动的部分

- `bazi.Calculate(...)`（`engine.go:182`）签名与逻辑不变。
- `POST /api/bazi/calculate`（`main.go:143`）不变。前端选定候选后，照常用「选中日期 + 时辰中点小时 + male/female」调用它。
- 入库模型、报告服务、所有下游分析不变。

## 前端设计

### 输入入口：首页「按生辰 / 按八字」切换

文件：`frontend/src/pages/HomePage.tsx`、`frontend/src/components/BirthProfileForm.tsx` 旁新增组件。

- 首页表单顶部加切换：`按生辰`（现有 `BirthProfileForm`）/ `按八字`（新面板）。默认「按生辰」。
- 新面板 `frontend/src/components/PillarsInputForm.tsx`：
  - 4 个干支下拉（年柱 / 月柱 / 日柱 / 时柱），每个下拉只列 60 个合法组合（甲子…癸亥），杜绝错字与非法柱。
  - 性别切换（复用现有控件）。
  - 可选「大致年代」输入（如出生年代下拉或年龄区间），用于把 `min_year/max_year` 缩小，减少候选数；不填则用默认范围。

### 候选选择交互

- 提交 → `POST /api/bazi/resolve-pillars`。
- `candidates.length === 0` → 表单内提示「这组八字找不到对应的真实日期，请核对四柱」。
- `=== 1` → 直接用该候选调用 `calculate`，进入结果页（无需用户再点）。
- `> 1` → 弹层 / 列表展示每个候选「公历日期 + 参考年龄（如 1965-08-12 · 约 60 岁）」，用户点选后调用 `calculate`。

### API 客户端

文件：`frontend/src/lib/api.ts` 新增 `resolvePillars(input)` 调用 `/api/bazi/resolve-pillars`，类型与后端对齐；复用现有 `calculate` 不变。

## 边界与校验

- **单柱合法性**：下拉只列 60 合法组合，前端层面即保证。
- **跨柱自洽性**（月干↔年干 五虎遁、时干↔日干 五鼠遁）：由反查兜底——不自洽的四柱搜不到任何候选，返回空，前端提示核对。无需单独写跨柱校验器。
- **子时**：候选落在 23:00–01:00 时沿用现有早/晚子时逻辑，反查统一按晚子时（0 时）取中点。
- **搜索范围**：默认 1900–2030，定义为包内常量，可调。

## 测试

### 后端（`backend/pkg/bazi/resolve_pillars_test.go`）

1. **正反一致**：取若干已知公历生日，先用 `Calculate` 正推出四柱，再 `ResolvePillars` 反查，断言候选中包含原日期。
2. **多候选**：构造跨 60 年仍成立的四柱，断言返回 ≥2 个候选且年份相差约 60。
3. **非法/不自洽**：传入月干与年干不符的四柱，断言返回空切片。
4. **时辰中点**：断言候选 `Hour` 为对应时辰中点（如时支为午 → Hour=12）。

### 后端 handler（`backend/internal/handler/bazi_handler_test.go`）

5. 缺省范围生效；空候选返回 200 + 空数组而非错误。

### 前端（`frontend/tests/`）

6. 「按八字」面板渲染 4 个干支下拉 + 性别。
7. 0 候选时显示核对提示；多候选时渲染候选列表。

## 非目标（YAGNI）

- 不做「纯八字降级模式」（无大运），已在方案选型中排除。
- 不做八字 OCR / 图片识别——用户自行从图片读出四柱填入下拉。
- 不缓存反查结果、不入库 resolve 中间态。
