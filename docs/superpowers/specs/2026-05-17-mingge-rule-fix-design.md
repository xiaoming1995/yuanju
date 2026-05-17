# 命格定格规则修正（方案 B）设计文档

**日期**：2026-05-17
**模块**：`backend/pkg/bazi/mingge.go`
**改动范围**：算法修正（无 API 变更、无 DB 变更、无前端变更）
**风险面**：会改变历史盘的格名输出 —— 属于期望中的纠错变化

---

## 1. 背景

现有 `DetectMingGe` 实现的是「七优先级透干取格法」，与命理师认可的标准规则相比，存在 4 处实质偏差：

| 偏差 | 规则要求 | 现实实现 |
|---|---|---|
| ① 第 1 层 | 月支透月干（主中余三气） | 仅主气透月干 |
| ③ 第 3 层 | 它支透月干（主中余三气） | 仅他支主气透月干 |
| ⑥ 第 6 层 | **克日干**的三合 / 三会成局立格 | 任意五行合局都立格（不判克日主） |
| ④ 第 7 层 | 频次≥4、吉凶同透**以凶神成格** | 频次≥4 取最高频，不分吉凶神 |

> 另有一处偏差 ②：第 5 层「多透中格神气壮」被压扁到第 2 / 4 层内做。这一处影响面大、需配套回归案例，本方案**不处理**，单独立项。

本方案目标：仅修上述 ①③④⑥ 四处，遵循 YAGNI，对前端 / 服务层 / DB 零改动。

---

## 2. 关键决策

经讨论确认：

| # | 决策点 | 决议 |
|---|---|---|
| A | 第 3 层遍历顺序 | 「主气优先于一切」：先扫三支主气，再扫三支中气，最后三支余气 |
| B | 第 6 层立格阴阳代表 | 沿用现行 `wuxingMainGan`（阳干代表）→ 立格为七杀格 / 偏印格 / 偏财格等 |
| C | 第 7 层凶神范围 | 严格 4 凶：七杀 / 伤官 / 偏印 / 劫财 |
| C2 | 第 7 层凶神内部决胜 | 七杀 > 伤官 > 偏印 > 劫财（按传统凶性递减） |
| C3 | 第 7 层吉神内部决胜 | 正官 > 正印 > 正财 > 食神 > 偏财 > 比肩 |
| ① 默认 | 月支三气透月干同盘命中 | 按主→中→余决胜（藏干表自然顺序，循环遇首匹配即返） |

---

## 3. 架构

依然是单文件 `mingge.go` 内的 `DetectMingGe`，无新文件、无接口变更。

```
DetectMingGe(r *BaziResult)
   │
   ├─ 第 1 层  月支三气 → 月干        [修复 ①]
   ├─ 第 2 层  月支任气 → 它干（保持现状）
   ├─ 第 3 层  他支三气 → 月干        [修复 ③]
   ├─ 第 4 层  他支任气 → 任干（保持现状）
   ├─ 第 5 层  留空（保持现状，本期不处理 ②）
   ├─ 第 6 层  合化局克日主 → 立格    [修复 ⑥ 新增 isKeWuxing]
   ├─ 第 7 层  全局频次 ≥ 4 → 凶神优先 → 内部决胜  [修复 ④]
   └─ 兜底     杂气格（保持现状）
```

新增 3 个文件级常量 / helper：

- `wuxingKe map[string]string` —— 五行相克查表
- `isKeWuxing(attacker, defender string) bool`
- `xiongShenSet map[string]bool` —— 凶神集合
- `xiongShenOrder []string` —— 凶神决胜顺序
- `jiShenOrder []string` —— 吉神决胜顺序

---

## 4. 详细改动

### 4.1 第 1 层 · 月支三气透月干

替换 `mingge.go` 第 166-173 行：

```go
// 第 1 层：月支主 / 中 / 余气 任一透月干（命中即停 → 主气优先）
if len(monthHideGans) > 0 {
    for _, hg := range monthHideGans {  // 已按 主→中→余 排序
        if hg == monthGan {
            ss := GetShiShen(dayGan, hg)
            geName := shiShenToGeName(ss)
            return geName, minggeDescDict[geName]
        }
    }
}
```

要点：`zhiHideGanFull` 表中藏干顺序天然是「主→中→余」，`for` 循环遇到第一个匹配即返回，无须额外排序。

### 4.2 第 3 层 · 他支三气透月干

替换第 207-218 行（按决策点 A：「先按主中余气」遍历）：

```go
// 第 3 层：年/日/时支 主 / 中 / 余气 任一透月干
//   按 主气 → 中气 → 余气 顺序扫描他支，命中即停
//   所有支最多 3 个藏干，所以外层固定 depth < 3
for depth := 0; depth < 3; depth++ {
    for _, oz := range otherZhis {
        hgs := zhiHideGanFull[oz]
        if depth >= len(hgs) {
            continue
        }
        if hgs[depth] == monthGan {
            ss := GetShiShen(dayGan, hgs[depth])
            geName := shiShenToGeName(ss)
            return geName, minggeDescDict[geName]
        }
    }
}
```

外层 `depth=0`（主气）→ 1（中气）→ 2（余气），内层遍历三支。这样**所有支的主气**先被扫完，才扫中气，才扫余气 —— 落实决策 A。

### 4.3 第 6 层 · 合化局克日主才立格

文件级新增：

```go
// 五行相克：木克土、土克水、水克火、火克金、金克木
var wuxingKe = map[string]string{
    "木": "土",
    "土": "水",
    "水": "火",
    "火": "金",
    "金": "木",
}

// isKeWuxing 判断 attacker 是否克 defender
func isKeWuxing(attacker, defender string) bool {
    return wuxingKe[attacker] == defender
}
```

替换第 256-262 行：

```go
// 第 6 层：无格时，克日干的三合 / 三会局立格
if wx := detectSanHeHui(allZhis); wx != "" {
    dayWx := ganWuxingMap[dayGan]
    if isKeWuxing(wx, dayWx) {
        if repGan, ok := wuxingMainGan[wx]; ok {
            ss := GetShiShen(dayGan, repGan)
            geName := shiShenToGeName(ss)
            return geName, minggeDescDict[geName]
        }
    }
}
```

非克日主的合局 → 继续往下走第 7 层 → 大多数情况下落到杂气格（如果第 7 层也不命中）。

### 4.4 第 7 层 · 吉凶神决胜

文件级新增：

```go
// 凶神（严格 4 凶口径）
var xiongShenSet = map[string]bool{
    "七杀": true, "伤官": true, "偏印": true, "劫财": true,
}

// 凶神内部决胜顺序：七杀 > 伤官 > 偏印 > 劫财
var xiongShenOrder = []string{"七杀", "伤官", "偏印", "劫财"}

// 吉神内部决胜顺序：正官 > 正印 > 正财 > 食神 > 偏财 > 比肩
var jiShenOrder = []string{"正官", "正印", "正财", "食神", "偏财", "比肩"}

// tieBreak 在频次并列的候选中按内部顺序决胜
func tieBreak(tiedShiShens []string) string {
    // 先按凶神顺序
    for _, x := range xiongShenOrder {
        for _, s := range tiedShiShens {
            if s == x {
                return s
            }
        }
    }
    // 再按吉神顺序
    for _, j := range jiShenOrder {
        for _, s := range tiedShiShens {
            if s == j {
                return s
            }
        }
    }
    // 兜底：返回第一个
    if len(tiedShiShens) > 0 {
        return tiedShiShens[0]
    }
    return ""
}
```

替换第 265-291 行：

```go
// 第 7 层：全局十神频次 ≥ 4 取最高频；吉凶同透以凶神成格
allSources := append([]string{}, allGans...)
for _, z := range allZhis {
    hgs := zhiHideGanFull[z]
    if len(hgs) > 0 {
        allSources = append(allSources, hgs[0]) // 仅主气
    }
}
shishenCount := make(map[string]int)
for _, g := range allSources {
    ss := GetShiShen(dayGan, g)
    if ss != "" {
        shishenCount[ss]++
    }
}

// 收集频次 ≥ 4 的所有十神
type cand struct {
    ss    string
    cnt   int
    xiong bool
}
var cands []cand
for ss, cnt := range shishenCount {
    if cnt >= 4 {
        cands = append(cands, cand{ss, cnt, xiongShenSet[ss]})
    }
}

if len(cands) > 0 {
    // 步骤 1：凶神优先 → 池子里如果有凶神就只在凶神中选；没凶神则在吉神中选
    var pool []cand
    for _, c := range cands {
        if c.xiong {
            pool = append(pool, c)
        }
    }
    if len(pool) == 0 {
        pool = cands // 全是吉神
    }

    // 步骤 2：池内按频次降序取最高
    maxCnt := pool[0].cnt
    for _, c := range pool[1:] {
        if c.cnt > maxCnt {
            maxCnt = c.cnt
        }
    }

    // 步骤 3：频次并列 → 按内部决胜顺序选
    var tied []string
    for _, c := range pool {
        if c.cnt == maxCnt {
            tied = append(tied, c.ss)
        }
    }
    pickedSS := tieBreak(tied)
    if pickedSS != "" {
        geName := shiShenToGeName(pickedSS)
        return geName, minggeDescDict[geName]
    }
}
```

---

## 5. 测试策略

新增 `backend/pkg/bazi/mingge_test.go`，覆盖每个修复点至少 1-2 个 case。

测试形式：直接构造 `BaziResult` 结构体，调 `DetectMingGe()`，断言返回的 `geName`。无需起公历盘。

### 5.1 Case 表

> **测试形式说明**：测试直接构造 `BaziResult` 结构体（手填年月日时柱字段），不需要起真实公历盘。表中"八字"是该 case 必须满足的判定条件，不一定对应某个真实出生日。实施阶段（plan）会精确填出每个 case 的 BaziResult 字段。

| Case | 判定条件 | 期望格 | 验证 |
|---|---|---|---|
| C1 月支主气透月干 | 月支寅 + 月干甲（寅主气=甲） | 建禄格 | 第 1 层 · 主气 |
| C2 月支中气透月干（修 ① 核心） | 月支寅 + 月干丙（寅中气=丙）+ 日干非丙 | 立月干十神对应格 | 第 1 层 · 中气 |
| C3 月支余气透月干（修 ① 核心） | 月支寅 + 月干戊（寅余气=戊）+ 日干非戊 | 立月干十神对应格 | 第 1 层 · 余气 |
| C4 他支主气透月干 | 他支主气 == 月干、月支不出该字 | 立月干十神对应格 | 第 3 层 · 主气 |
| C5 他支中气透月干（修 ③ 核心） | 他支中气 == 月干、月支不出、他支主气不出 | 立月干十神对应格 | 第 3 层 · 中气 |
| C6 三合金局 + 甲日（修 ⑥ 通过）| 甲日 + 巳/酉/丑 三支齐 + 无透干立格 | 七杀格 | 第 6 层 · 立格 |
| C7 三合水局 + 甲日（修 ⑥ 关键过滤）| 甲日 + 申/子/辰 三支齐 + 无透干立格 + 第 7 层亦不命中 | **杂气格**（不应立偏印格） | 第 6 层 · 过滤 |
| C8 三会木局 + 甲日 | 甲日 + 寅/卯/辰 三支齐 + 无透干立格 + 第 7 层亦不命中 | **杂气格**（同为木比劫，不克日）| 第 6 层 · 过滤 |
| C9 4 吉 + 4 凶并列（修 ④ 核心）| 构造一盘使伤官计数 = 食神计数 = 4 | 伤官格 | 第 7 层 · 凶优先 |
| C10 4 凶并列内部决胜（修 ④）| 构造一盘使七杀 = 伤官 = 4 | 七杀格 | 第 7 层 · 凶神内部 |
| C11 4 吉并列内部决胜 | 构造一盘使正官 = 食神 = 4 | 正官格 | 第 7 层 · 吉神内部 |
| C12 不命中任何层 | 各神频次 ≤ 3、无合局、无透干 | 杂气格 | 兜底 |

### 5.2 回归测试

`go test ./pkg/bazi/...` 全跑。预期：
- 现有 mingge 相关测试可能有部分用例的期望需要更新（属于纠错，不是 regression）
- 其它非 mingge 测试（shensha、tiaohou、yongshen 等）**不应受影响**

---

## 6. 数据流 / 调用方影响

```
CalculateBazi (engine.go)
    │
    ├─ DetectMingGe(r) ──→ 返回值结构不变（name, desc）
    │      ↑
    │      仅算法行为变化
    │
    └─ r.MingGe / r.MingGeDesc 字段不变
              │
              ▼
        report_service.go        ← 不感知
        ResultPage / PDF         ← 不感知
        ai_polished_reports      ← 不感知
```

零接口变更、零 DDL 变更。**算法纠错型改动，对外不可见**。

---

## 7. 验收标准

1. `go test ./pkg/bazi/...` 全绿
2. 新增 mingge_test.go 中 12 个 case 全过
3. 现有非 mingge 测试无回归
4. `go build ./...` 通过

---

## 8. 范围外 / 未来工作

- **偏差 ②**：第 5 层「多透中格神气壮」跨层综合 —— 单独立项，需先准备回归案例集（≥ 30 盘）
- 第 6 层「合化破格」判定（合化所成的格神被合走 / 冲克）
- 透干位置加权（月干 > 时干 > 年干/日干）
- 月令司令深浅加权（本气 vs 中气 vs 余气在天数上的分配）
- 格局描述动态化（结合身强弱 / 用神到位 / 制化判定）

以上均为后续优化路径，本期均不处理。
