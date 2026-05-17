# 命格取格规则·命师口径设计文档

**日期**：2026-05-17
**模块**：`backend/pkg/bazi/mingge.go`（整体重写 `DetectMingGe`）
**改动范围**：算法整体重写。无 API 变更，无 DB 变更，无前端变更。
**风险面**：会改变历史盘的格名输出 —— 属于命师对齐型纠错。

---

## 1. 背景

前作 [`2026-05-17-mingge-rule-fix-design.md`](./2026-05-17-mingge-rule-fix-design.md)（方案 B）按"教科书七优先级表"修补现有算法。实施完毕后用 12 个专业命师标注的真实盘对照，仅 **5/12** 通过。深入分析揭示：**命师的实际取格方法根本不基于七优先级分层**，而是基于「四柱透干 + 通根强度评分 + 几条隐性优先级规则」的整体评估。

本方案：**整体重写** `DetectMingGe`，实现命师 6 条隐性规则。验证集 **12/12** 通过。

---

## 2. 命师 6 条隐性规则

### 规则 0：月禄格 / 月刃格 special case

- **月禄格**：月支正好是日干的**临官**位 → 立月禄格，跳过所有其他判断
- **月刃格**：**阳干日**（甲丙戊庚壬）+ 月支正好是日干的**帝旺**位 → 立月刃格，跳过所有其他判断

> 注：阴干日的帝旺位不立"月刃"。"羊刃"传统专指阳干。

十二长生临官/帝旺表：

| 日干 | 临官（月禄）| 帝旺（月刃，仅阳干）|
|---|---|---|
| 甲 | 寅 | 卯 |
| 乙 | 卯 | （阴干无月刃）|
| 丙 / 戊 | 巳 | 午 |
| 丁 / 己 | 午 | （阴干无月刃）|
| 庚 | 申 | 酉 |
| 辛 | 酉 | （阴干无月刃）|
| 壬 | 亥 | 子 |
| 癸 | 子 | （阴干无月刃）|

### 规则 1：候选 = 四柱天干中非比劫的十神

- 收集年干、月干、时干（日干自身**不计**）相对日干的十神
- **剔除**比肩 / 劫财（除非已被规则 0 触发立月禄/月刃）

> 注：日干自身 = 比肩，但日干不参与"候选"评估。

### 规则 4：食 / 伤同透 → 强制立伤官格

- 候选集合中**同时存在**食神 + 伤官时 → 立**伤官格**
- **此条无视通根** —— 即使伤官透出且无任何地支根，也立伤官格

> 解释：传统命师认为"食伤齐透时伤官更显格象，食神隐而伤官显"。

### 规则 3：按通根强度排序候选，取最强者立格

通根强度评分：

| 通根位置 | 评分 |
|---|---|
| 月支主气 | 6 |
| 月支中气 | 5 |
| 月支余气 | 4 |
| 他支主气 | 3 |
| 他支中气 | 2 |
| 他支余气 | 1 |
| 无根 | 0 |

- 每个候选取它在四柱地支藏干中能找到的**最强单根**作为该候选的通根分
- 候选按通根分降序排序；分高者立格
- 同分时优先级：规则 0 → 规则 4 → 规则 3 内部决胜（同分内部留待 future work，先简单取首个）

### 规则 6：地支气势集中于"财"五行 → 立财格

触发条件（**所有条件 AND**）：

1. 规则 0、4 都未触发
2. 规则 3 的候选要么为空，要么**全部无根**（最强通根分 = 0）
3. 四柱**地支主气**同属一个五行 X
4. X 是日干的**财**五行（日干克 X，即 wuxingKe[日干五行] == X）

满足时：取**月支主气**配日干算十神，立其格（结果应为偏财格或正财格）。

### 规则 5：兜底 — 杂气格

以上规则都不触发 → 立**杂气格**（命师口径 = "不成格"）。

---

## 3. 算法流程

```
DetectMingGe(r *BaziResult) (geName, desc string):

  ─── 0. 月禄 / 月刃 special case ───
  if month_zhi == 临官位(day_gan):
      return "月禄格", desc

  if 阳干(day_gan) and month_zhi == 帝旺位(day_gan):
      return "月刃格", desc

  ─── 1. 收集候选（天干上非比劫十神）───
  candidates = []
  for gan in [year_gan, month_gan, hour_gan]:           # 注：不含 day_gan 自身
      shishen = GetShiShen(day_gan, gan)
      if shishen in ("比肩", "劫财"):
          continue
      candidates.append((gan, shishen))

  ─── 4. 食/伤同透优先 ───
  has_food = any(s == "食神" for _, s in candidates)
  has_injury = any(s == "伤官" for _, s in candidates)
  if has_food and has_injury:
      return "伤官格", desc

  ─── 3. 通根评分排序，取最强 ───
  if candidates:
      sort candidates by root_strength(gan) descending
      best_gan, best_shishen = candidates[0]
      best_root = root_strength(best_gan)
      if best_root > 0:
          return shishen_to_ge_name(best_shishen), desc

  ─── 6. 地支气势集中于财 → 立财格 ───
  zhi_main_wuxings = [
      gan_wuxing[zhi_hide_gan_full[year_zhi][0]],
      gan_wuxing[zhi_hide_gan_full[month_zhi][0]],
      gan_wuxing[zhi_hide_gan_full[day_zhi][0]],
      gan_wuxing[zhi_hide_gan_full[hour_zhi][0]],
  ]
  if all 4 same wuxing X:
      day_wx = gan_wuxing[day_gan]
      if isKeWuxing(day_wx, X):       # 日干克 X => X 是财
          month_main_gan = zhi_hide_gan_full[month_zhi][0]
          ss = GetShiShen(day_gan, month_main_gan)
          return shishen_to_ge_name(ss), desc

  ─── 5. 兜底 ───
  return "杂气格", desc


root_strength(gan):
  strength = 0
  for (pos, zhi) in [("month", month_zhi), ("year", year_zhi),
                     ("day", day_zhi), ("hour", hour_zhi)]:
      hgs = zhi_hide_gan_full[zhi]
      for i, hg in enumerate(hgs):       # i=0 主气, i=1 中气, i=2 余气
          if hg == gan:
              if pos == "month":
                  s = 6 - i              # 主气=6, 中气=5, 余气=4
              else:
                  s = 3 - i              # 主气=3, 中气=2, 余气=1
              if s > strength:
                  strength = s
  return strength
```

---

## 4. 验证集（12 case · 命师标注）

| # | 八字（公历）| 性别 | 四柱（年月日时）| 命师 | 触发规则 |
|---|---|---|---|---|---|
| C1 | 1995-10-12 11时 | 男 | 乙亥 丙戌 丙子 甲午 | 偏印格 | 规则 3（甲 在亥中气根=2，唯一候选）|
| C2 | 1996-02-08 20时 | 男 | 丙子 庚寅 乙亥 丙戌 | 伤官格 | 规则 3（丙 在寅中气=5，月支中气根强 > 庚无根）|
| C3 | 1991-02-07 16时 | 女 | 辛未 庚寅 戊申 庚申 | 伤官格 | 规则 4（食伤同透取伤）|
| C4 | 1997-12-01 12时 | 女 | 丁丑 辛亥 丁丑 丙午 | 偏财格 | 规则 3（辛 在丑余气根=1，唯一候选）|
| C5 | 1988-01-18 04时 | 女 | 丁卯 癸丑 壬申 壬寅 | 杂气格（不成格）| 规则 5（丁无根，单候选）|
| C6 | 1996-11-08 08时 | 女 | 丙子 己亥 己酉 戊辰 | 杂气格（不成格）| 规则 5（丙无根）|
| C7 | 1991-12-30 10时 | 女 | 辛未 庚子 甲戌 己巳 | 正财格 | 规则 3（己 在未主气=3 > 庚中气=2 > 辛余气=1）|
| C8 | 1996-12-16 22时 | 男 | 丙子 庚子 丁亥 辛亥 | 杂气格（不成格）| 规则 5（庚辛无根 + 地支水非财）|
| C9 | 1995-01-23 16时 | 男 | 甲戌 丁丑 甲寅 壬申 | 偏印格 | 规则 3（壬 申中气=2 > 丁戌余气=1）|
| C10 | 1993-01-16 14时 | 女 | 壬申 癸丑 丁酉 丁未 | 七杀格 | 规则 3（癸 月支中气=5 > 壬他支中气=2）|
| C11 | 2015-02-02 18时40分 | 男 | 甲午 丁丑 己酉 癸酉 | 偏财格 | 规则 3（癸 月支中气=5 > 丁主气=3 > 甲无根）|
| C12 | 1985-04-26 19时46分 | 男 | 乙丑 庚辰 乙未 丙戌 | 正财格（命师笼统称"财格"；月支主气=戊，乙阴克戊阳 异性 → 正财）| 规则 6（4 支主气全土 = 乙日财；按月支主气 戊 → 正财格）|

> C11、C12 含分钟数；测试用 hour 部分（与生产 `bazi.Calculate(year, month, day, hour, ...)` 一致，不支持分钟）。

---

## 5. 实施细节

### 5.1 现有 `mingge.go` 处理

**整体重写** `DetectMingGe` 函数。保留：
- `zhiHideGanFull` 表（藏干主中余顺序，新算法依赖）
- `sanHeJu` / `sanHuiJu`（暂不删，备未来用；新算法不引用）
- `wuxingMainGan` / `ganWuxingMap` / `wuxingKe` 表（沿用）
- `isKeWuxing` 函数（沿用）
- `shiShenToGeName` 函数（沿用，但建禄格/月刃格命名只在规则 0 走通时使用；规则 1 跳过比劫导致它们不会自动产出）
- `minggeDescDict` 字典（沿用）
- `detectSanHeHui` 函数（**新算法不调用**，但保留方便单测）

**删除**：
- 七优先级的 Layer 1-7 主干逻辑（替换成新算法）
- `gansContains` / `countGanInGans` / `wuxingScore` / `candidate struct`（旧算法 2/4 层用）
- 新增的 `xiongShenSet` / `xiongShenOrder` / `jiShenOrder` / `tieBreak`（旧算法 layer 7 用）—— 注意：前作 `feat/mingge-rule-fix-b` 分支引入了这些，但 main 上不存在，无需处理

### 5.2 新增

- `linGuanZhi`：日干 → 临官地支 map（10 干 → 1 地支）
- `diWangZhi`：日干 → 帝旺地支 map（10 干 → 1 地支）
- `yangGans`：阳干 set（甲丙戊庚壬）
- `rootStrength(gan, r)`：按上述评分表算单个天干的最强通根分
- 整体重写 `DetectMingGe`
- **`minggeDescDict` 新增条目 `"月禄格"`**（当前字典只有"建禄格"；月禄格在传统命名上即建禄格的同义词，使用同样的释义文字即可）

### 5.3 数据流不变

```
CalculateBazi (engine.go)
   ↓
DetectMingGe(r) → (name, desc string)   ← 接口不变
   ↓
r.MingGe / r.MingGeDesc                  ← 字段不变
   ↓
report_service / PDF / 前端              ← 不感知
```

---

## 6. 测试策略

### 6.1 文件

- 替换 `backend/pkg/bazi/mingge_test.go`（main 上还没有这个文件 —— `feat/mingge-rule-fix-b` 上有但已废分支）

### 6.2 12 case 验证

每个 case 表驱动测试，断言 `geName` 等于命师标注值（C5/C6/C8 = "杂气格"）。

### 6.3 加 helper 单测

- `TestLinGuanDiWang`：验证 临官/帝旺表对全 10 干正确
- `TestRootStrength_MonthBranch`：验证月支主气根=6、月支中气=5、月支余气=4
- `TestRootStrength_OtherBranch`：验证他支主气=3、中气=2、余气=1
- `TestRootStrength_NoRoot`：返回 0

### 6.4 回归 / 兼容

跑 `go test ./...` 全绿。已有的 `internal/service/report_service_test.go:254` 用 1996-02-08 20时（即 C2），命师标注为伤官格。新算法预期也输出伤官格 —— 不应造成回归。

---

## 7. 验收

1. `go test ./pkg/bazi/...` 全绿
2. `mingge_test.go` 12 个 case 全过
3. `go test ./...` 无回归
4. `go build ./...` / `go vet ./...` 清

---

## 8. 范围外 / 未来工作

- **规则 3 同分内部决胜**：当前实现"同分取首个"。未来可加位置权重（月柱优先于他柱、时柱次之）。但当前 12 case 中无此场景需求。
- **规则 6 扩展**：本方案仅覆盖"地支气势全土 = 财"。命师可能对其他类象格（曲直 / 稼穑 / 从杀 等）有判断 —— 当前 12 case 中只有 C12 一例，无法泛化。如未来增加 case，可扩展。
- **食伤同透取伤的边界**：当前规则 4 是"二者都透则取伤"。如果只透食神无伤官、伤官孤透有根，则正常走规则 3 —— 这是预期行为。
- **月禄/月刃下的具体十神描述**：当前用 `minggeDescDict["月禄格"]`、`minggeDescDict["月刃格"]`，需要确认字典里有这两条；如果没有，新增或复用"建禄格"/"月刃格"的描述。
- **规则 0 与规则 6 冲突时的优先级**：理论上月禄格触发后不会再到规则 6。本设计明确规则 0 早 return，无冲突。
