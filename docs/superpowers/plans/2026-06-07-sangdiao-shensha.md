# 丧吊类神煞（丧门·披麻·三丘·五墓）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在个人命盘四柱、大运、流年中新增「丧门、披麻、三丘、五墓」四个丧吊类凶煞。

**Architecture:** 神煞只有两个计算入口——`GetPillarsShenSha`（四柱→个人报告）与 `GetDayunShenSha`（运柱→大运，且被流年复用）。在这两个函数中追加四个神煞的计算；三丘/五墓的「月支→目标干支」映射抽成两个包级 helper 供两处共用；极性写入 `ShenShaPolarity`；释义通过新 goose migration 写入 `shensha_annotations`。合盘不动。

**Tech Stack:** Go 1.25（`pkg/bazi` 纯函数）、goose migrations（embed FS）、Postgres `shensha_annotations` 表。

**设计依据：** `docs/superpowers/specs/2026-06-07-sangdiao-shensha-design.md`

**关键约定（实现前必读）：**
- 丧门/披麻：年支起例，**只比地支**（与既有「吊客」同源）。丧门=年支+2，披麻=年支−3。
- 三丘/五墓：月支起例，**比全干支（天干+地支）**（用户选定 B 方案，与既有魁罡/十恶大败同档）。三丘按季节，五墓按月（季月辰未戌丑→戊辰）。
- 四柱扫描三丘/五墓时**跳过月柱（索引 1）**（基准柱自身不参与，对应 spec §2.5）。大运/流年不涉及月柱，不跳过。
- 现有代码惯例：`diaokeMap` 在 `GetPillarsShenSha` 与 `GetDayunShenSha` **各自重复定义**（不共享）。本计划遵循该惯例——`sangmenMap`/`pimaMap` 两处各写一份；仅把含 switch 逻辑的 `sanqiuGanZhi`/`wumuGanZhi` 抽成共享 helper（避免季节判断逻辑出现第三份）。

---

## File Structure

| 文件 | 职责 | 改动 |
|---|---|---|
| `backend/pkg/bazi/shensha.go` | 四柱神煞 + 极性表 + 三丘/五墓共享 helper | 加 4 极性行；加 2 个包级 helper；`GetPillarsShenSha` 末尾加 4 神煞计算 |
| `backend/pkg/bazi/shensha_dayun.go` | 大运/流年神煞 | `GetDayunShenSha` 末尾加 4 神煞计算（复用 helper）|
| `backend/pkg/bazi/shensha_sangdiao_test.go` | 本特性的断言测试 | 新建 |
| `backend/pkg/database/migrations/00015_add_sangdiao_shensha_annotations.sql` | 4 个神煞释义 seed | 新建 |

---

## Task 1: 四柱（GetPillarsShenSha）+ 极性 + 共享 helper

**Files:**
- Create: `backend/pkg/bazi/shensha_sangdiao_test.go`
- Modify: `backend/pkg/bazi/shensha.go`（`ShenShaPolarity` 凶煞区；新增 2 helper；`GetPillarsShenSha` 末尾，现有「第八组：天罗地网」块之后、`return result` 之前）

- [ ] **Step 1: 写失败测试**

创建 `backend/pkg/bazi/shensha_sangdiao_test.go`：

```go
package bazi

import (
	"slices"
	"testing"
)

// GetPillarsShenSha 参数顺序：yg, yz, mg, mz, dg, dz, hg, hz
// 返回 [4][]string，索引 0=年 1=月 2=日 3=时

func TestPillars_SangMen(t *testing.T) {
	// 年支子 → 丧门=寅；月支寅 → 月柱(索引1)命中
	r := GetPillarsShenSha("甲", "子", "丙", "寅", "戊", "辰", "庚", "午")
	if !slices.Contains(r[1], "丧门") {
		t.Fatalf("月柱应含丧门，实际 %v", r[1])
	}
}

func TestPillars_PiMa(t *testing.T) {
	// 年支子 → 披麻=酉；时支酉 → 时柱(索引3)命中
	r := GetPillarsShenSha("甲", "子", "丁", "卯", "己", "巳", "癸", "酉")
	if !slices.Contains(r[3], "披麻") {
		t.Fatalf("时柱应含披麻，实际 %v", r[3])
	}
}

func TestPillars_SanQiu_SpringHit(t *testing.T) {
	// 月支寅(春) → 三丘=辛丑；年柱=辛丑(非月柱) → 索引0命中
	r := GetPillarsShenSha("辛", "丑", "庚", "寅", "戊", "戌", "甲", "午")
	if !slices.Contains(r[0], "三丘") {
		t.Fatalf("年柱应含三丘，实际 %v", r[0])
	}
}

func TestPillars_SanQiu_StrictGanZhi(t *testing.T) {
	// 月支寅(春) → 三丘=辛丑；年柱=己丑(地支对、天干错) → 不命中（B 方案严格性）
	r := GetPillarsShenSha("己", "丑", "庚", "寅", "戊", "戌", "甲", "午")
	if slices.Contains(r[0], "三丘") {
		t.Fatalf("己丑≠辛丑，年柱不应含三丘，实际 %v", r[0])
	}
}

func TestPillars_WuMu_SeasonMonthHitNonMonthPillar(t *testing.T) {
	// 月支辰(季月) → 五墓=戊辰；日柱=戊辰(非月柱) → 索引2命中
	r := GetPillarsShenSha("甲", "子", "庚", "辰", "戊", "辰", "丙", "午")
	if !slices.Contains(r[2], "五墓") {
		t.Fatalf("日柱应含五墓，实际 %v", r[2])
	}
}

func TestPillars_WuMu_ExcludeMonthSelfMatch(t *testing.T) {
	// 月支辰(季月) → 五墓=戊辰；月柱本身=戊辰，但月柱为基准柱，应排除
	r := GetPillarsShenSha("甲", "子", "戊", "辰", "乙", "酉", "丙", "午")
	if slices.Contains(r[1], "五墓") {
		t.Fatalf("月柱自命中应被排除，实际 %v", r[1])
	}
}

func TestPillars_WuMu_OrdinaryMonth(t *testing.T) {
	// 月支寅(正月) → 五墓=乙未；日柱=乙未 → 索引2命中
	r := GetPillarsShenSha("甲", "子", "丙", "寅", "乙", "未", "庚", "午")
	if !slices.Contains(r[2], "五墓") {
		t.Fatalf("日柱应含五墓，实际 %v", r[2])
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run 'TestPillars_(SangMen|PiMa|SanQiu|WuMu)' -v`
Expected: 全部 FAIL（神煞未实现，断言不通过）。

- [ ] **Step 3: 加极性表 4 行**

在 `backend/pkg/bazi/shensha.go` 的 `ShenShaPolarity` map「凶煞」区（现有 `"吊客": "xiong",` / `"墓门": "xiong",` 附近）追加：

```go
	"丧门":   "xiong",
	"披麻":   "xiong",
	"三丘":   "xiong",
	"五墓":   "xiong",
```

- [ ] **Step 4: 加两个共享 helper**

在 `backend/pkg/bazi/shensha.go` 中（建议紧接 `GetPillarsShenSha` 函数之后、文件内合适位置），新增：

```go
// sanqiuGanZhi 返回某月支(按季节)对应的「三丘」目标干支；未知月支返回 ""。
// 春(寅卯辰)→辛丑、夏(巳午未)→壬辰、秋(申酉戌)→乙未、冬(亥子丑)→丙戌。
func sanqiuGanZhi(monthZhi string) string {
	switch {
	case strings.Contains("寅卯辰", monthZhi):
		return "辛丑"
	case strings.Contains("巳午未", monthZhi):
		return "壬辰"
	case strings.Contains("申酉戌", monthZhi):
		return "乙未"
	case strings.Contains("亥子丑", monthZhi):
		return "丙戌"
	}
	return ""
}

// wumuGanZhi 返回某月支对应的「五墓」目标干支；四季月(辰未戌丑)→戊辰；未知月支返回 ""。
// 寅卯→乙未、巳午→丙戌、申酉→辛丑、亥子→壬辰、辰未戌丑→戊辰。
func wumuGanZhi(monthZhi string) string {
	switch monthZhi {
	case "寅", "卯":
		return "乙未"
	case "巳", "午":
		return "丙戌"
	case "申", "酉":
		return "辛丑"
	case "亥", "子":
		return "壬辰"
	case "辰", "未", "戌", "丑":
		return "戊辰"
	}
	return ""
}
```

> `strings` 已是 `shensha.go` 的现有 import，无需新增。

- [ ] **Step 5: 在 GetPillarsShenSha 末尾加计算**

在 `backend/pkg/bazi/shensha.go` 的 `GetPillarsShenSha` 函数内，现有「第八组：天罗地网」块之后、`return result` 之前插入：

```go
	// ══════════════════════════════════════════════════════════
	// 第九组：丧吊类之丧门/披麻（年支起例，地支匹配，与吊客同源）
	// 丧门=年支+2位, 披麻=年支-3位
	// ══════════════════════════════════════════════════════════
	sangmenMap := map[string]string{
		"子": "寅", "丑": "卯", "寅": "辰", "卯": "巳",
		"辰": "午", "巳": "未", "午": "申", "未": "酉",
		"申": "戌", "酉": "亥", "戌": "子", "亥": "丑",
	}
	if smZhi, ok := sangmenMap[yz]; ok {
		for i, z := range zhis {
			addIf(i, z == smZhi, "丧门")
		}
	}
	pimaMap := map[string]string{
		"子": "酉", "丑": "戌", "寅": "亥", "卯": "子",
		"辰": "丑", "巳": "寅", "午": "卯", "未": "辰",
		"申": "巳", "酉": "午", "戌": "未", "亥": "申",
	}
	if pmZhi, ok := pimaMap[yz]; ok {
		for i, z := range zhis {
			addIf(i, z == pmZhi, "披麻")
		}
	}

	// ══════════════════════════════════════════════════════════
	// 第十组：丧吊类之三丘/五墓（月支起例，全干支严格匹配）
	// 月柱(索引1)为基准柱，自身不参与匹配（基准柱自身不参与惯例）
	// ══════════════════════════════════════════════════════════
	if sqTarget := sanqiuGanZhi(mz); sqTarget != "" {
		for i, gz := range ganZhis {
			if i == 1 {
				continue
			}
			addIf(i, gz == sqTarget, "三丘")
		}
	}
	if wmTarget := wumuGanZhi(mz); wmTarget != "" {
		for i, gz := range ganZhis {
			if i == 1 {
				continue
			}
			addIf(i, gz == wmTarget, "五墓")
		}
	}
```

> `zhis`、`ganZhis`、`addIf` 均为 `GetPillarsShenSha` 已有的局部变量（见函数开头），直接复用。

- [ ] **Step 6: 跑测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run 'TestPillars_(SangMen|PiMa|SanQiu|WuMu)' -v`
Expected: 全部 PASS。

- [ ] **Step 7: 跑整包回归，确认未破坏既有神煞**

Run: `cd backend && go test ./pkg/bazi/...`
Expected: ok（含既有 `TestShenShaOutputAll` 等）。

- [ ] **Step 8: Commit**

```bash
git add backend/pkg/bazi/shensha.go backend/pkg/bazi/shensha_sangdiao_test.go
git commit -m "feat(bazi): add 丧门/披麻/三丘/五墓 to four-pillar shensha"
```

---

## Task 2: 大运/流年（GetDayunShenSha）

**Files:**
- Modify: `backend/pkg/bazi/shensha_dayun.go`（`GetDayunShenSha` 末尾，现有「天医」块之后、`return result` 之前）
- Modify: `backend/pkg/bazi/shensha_sangdiao_test.go`（追加大运测试）

- [ ] **Step 1: 追加失败测试**

在 `backend/pkg/bazi/shensha_sangdiao_test.go` 末尾追加：

```go
// GetDayunShenSha 参数顺序：yearGan, yearZhi, monthZhi, dayGan, dayZhi, dayunGan, dayunZhi

func TestDayun_SangMen(t *testing.T) {
	// 年支子 → 丧门=寅；大运支=寅 → 命中
	r := GetDayunShenSha("甲", "子", "寅", "戊", "辰", "丙", "寅")
	if !slices.Contains(r, "丧门") {
		t.Fatalf("大运应含丧门，实际 %v", r)
	}
}

func TestDayun_PiMa(t *testing.T) {
	// 年支子 → 披麻=酉；大运支=酉 → 命中
	r := GetDayunShenSha("甲", "子", "寅", "戊", "辰", "乙", "酉")
	if !slices.Contains(r, "披麻") {
		t.Fatalf("大运应含披麻，实际 %v", r)
	}
}

func TestDayun_SanQiu_Hit(t *testing.T) {
	// 月支寅(春) → 三丘=辛丑；大运=辛丑 → 命中
	r := GetDayunShenSha("甲", "子", "寅", "戊", "辰", "辛", "丑")
	if !slices.Contains(r, "三丘") {
		t.Fatalf("大运应含三丘，实际 %v", r)
	}
}

func TestDayun_SanQiu_StrictGanZhi(t *testing.T) {
	// 月支寅(春) → 三丘=辛丑；大运=己丑(地支对、天干错) → 不命中
	r := GetDayunShenSha("甲", "子", "寅", "戊", "辰", "己", "丑")
	if slices.Contains(r, "三丘") {
		t.Fatalf("己丑≠辛丑，大运不应含三丘，实际 %v", r)
	}
}

func TestDayun_WuMu_SeasonMonth(t *testing.T) {
	// 月支辰(季月) → 五墓=戊辰；大运=戊辰 → 命中（大运无月柱排除）
	r := GetDayunShenSha("甲", "子", "辰", "戊", "酉", "戊", "辰")
	if !slices.Contains(r, "五墓") {
		t.Fatalf("大运应含五墓，实际 %v", r)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run 'TestDayun_(SangMen|PiMa|SanQiu|WuMu)' -v`
Expected: 全部 FAIL。

- [ ] **Step 3: 在 GetDayunShenSha 末尾加计算**

在 `backend/pkg/bazi/shensha_dayun.go` 的 `GetDayunShenSha` 函数内，现有「天医」块之后、`return result` 之前插入：

```go
	// ── 丧吊类：丧门/披麻（年支起例，比运支）+ 三丘/五墓（月支起例，比运柱全干支）──
	sangmenMap := map[string]string{
		"子": "寅", "丑": "卯", "寅": "辰", "卯": "巳",
		"辰": "午", "巳": "未", "午": "申", "未": "酉",
		"申": "戌", "酉": "亥", "戌": "子", "亥": "丑",
	}
	if smZhi, ok := sangmenMap[yz]; ok {
		add(z == smZhi, "丧门")
	}
	pimaMap := map[string]string{
		"子": "酉", "丑": "戌", "寅": "亥", "卯": "子",
		"辰": "丑", "巳": "寅", "午": "卯", "未": "辰",
		"申": "巳", "酉": "午", "戌": "未", "亥": "申",
	}
	if pmZhi, ok := pimaMap[yz]; ok {
		add(z == pmZhi, "披麻")
	}
	if sqTarget := sanqiuGanZhi(mz); sqTarget != "" {
		add(g+z == sqTarget, "三丘")
	}
	if wmTarget := wumuGanZhi(mz); wmTarget != "" {
		add(g+z == wmTarget, "五墓")
	}
```

> `yz`、`mz`、`g`、`z`、`add` 均为 `GetDayunShenSha` 已有的局部变量（见函数开头）。`sanqiuGanZhi`/`wumuGanZhi` 为 Task 1 在同包 `shensha.go` 定义的 helper，直接调用。`sangmenMap`/`pimaMap` 在此重复定义，遵循同文件 `diaokeMap` 各自定义的既有惯例。

- [ ] **Step 4: 跑测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run 'TestDayun_(SangMen|PiMa|SanQiu|WuMu)' -v`
Expected: 全部 PASS。

- [ ] **Step 5: 整包回归**

Run: `cd backend && go test ./pkg/bazi/...`
Expected: ok。

- [ ] **Step 6: Commit**

```bash
git add backend/pkg/bazi/shensha_dayun.go backend/pkg/bazi/shensha_sangdiao_test.go
git commit -m "feat(bazi): add 丧门/披麻/三丘/五墓 to dayun/liunian shensha"
```

---

## Task 3: 释义 migration（shensha_annotations seed）

**Files:**
- Create: `backend/pkg/database/migrations/00015_add_sangdiao_shensha_annotations.sql`

- [ ] **Step 1: 创建 migration 文件**

创建 `backend/pkg/database/migrations/00015_add_sangdiao_shensha_annotations.sql`，内容如下（释义为凶煞、克制不渲染恐怖，文案中不含英文单引号，无需转义）：

```sql
-- +goose Up

INSERT INTO shensha_annotations (name, polarity, description) VALUES ('丧门', 'xiong', '丧门是以年支推算的凶煞，与吊客、披麻同属丧吊一族，旧说主孝服、哀伤与离别之事。命带丧门者，一生中与丧葬、探病、吊唁等场合的缘分较深，情绪上也较易因亲友的健康或聚散而牵动。此星落于不同柱位，对应不同人生阶段的际遇：年柱关乎早年家庭，月柱关乎中青年的亲缘往来，日柱关乎配偶与自身，时柱关乎晚景与子女。需说明的是，神煞只是参考信号，并非定数，命局整体的强弱喜忌才是关键。命带丧门更宜理解为「于生离死别之事较为敏感」，提醒在相关时节多关照长辈健康、稳住心绪即可，不必因星名生忧。') ON CONFLICT (name) DO NOTHING;
INSERT INTO shensha_annotations (name, polarity, description) VALUES ('披麻', 'xiong', '披麻是以年支推算的凶煞（年支后三位），取「披麻戴孝」之意，与丧门、吊客同主孝服哀戚之事。命带披麻者，较易遇到亲友丧病、聚散无常的场景，情绪上也容易染上一层淡淡的忧思。古人将其与丧门并称为孝丧之星，但它分量并不算重，更多是一种情绪与际遇上的提示，而非灾祸的判语。落于年/月/日/时不同柱位，分别对应早年、中年、自身与晚景的相关际遇。理性看待即可：命带披麻只说明此人对离别之事较为在意，遇相关流年宜多陪伴家人、留意长辈起居，不必因星名而生畏惧。') ON CONFLICT (name) DO NOTHING;
INSERT INTO shensha_annotations (name, polarity, description) VALUES ('三丘', 'xiong', '三丘是以出生季节推算的凶煞，与五墓相对成双，旧说主疾病、孝服与坟茔之事，属丧吊一族。其取法按四季各定一目标干支（春辛丑、夏壬辰、秋乙未、冬丙戌），四柱命中者即为带星。命带三丘者，传统上认为与六亲的健康、聚散较为相关，对生老病死之事也可能更为敏感。需强调的是，三丘属冷门小煞，分量轻，且各流派取法不一，只能作为参考信号，绝非凶险定论。命局的整体格局、五行喜忌远比单一神煞重要；理解为「于亲缘健康之事宜多留心」即可，不必据此生畏。') ON CONFLICT (name) DO NOTHING;
INSERT INTO shensha_annotations (name, polarity, description) VALUES ('五墓', 'xiong', '五墓是以出生月份推算的凶煞，与三丘相对成对，旧说主坟茔、疾病与孝服之事。其取法按月各定一目标干支（如正二月乙未、四五月丙戌、四季月戊辰等），四柱命中者为带星（作为基准的月柱本身不计）。命带五墓者，传统认为与六亲健康、家宅安宁略有牵连，重见者古书谓主骨肉刑伤，然此说偏重，实际更宜淡看。五墓与三丘同为冷门小煞，流派分歧大、分量轻，仅作参考信号之一。命运吉凶终究取决于格局与五行平衡，而非一二神煞；命带五墓，理解为「于亲缘与健康之事多一分留意」便好。') ON CONFLICT (name) DO NOTHING;
```

- [ ] **Step 2: 验证 migration 能干净应用（编译 + 既有迁移测试）**

Run: `cd backend && go build ./... && go test ./pkg/database/...`
Expected: build ok；migration 相关测试 ok（新文件被 `//go:embed migrations/*.sql` 自动纳入，goose 顺序应用至 00015）。

- [ ] **Step 3: Commit**

```bash
git add backend/pkg/database/migrations/00015_add_sangdiao_shensha_annotations.sql
git commit -m "feat(db): seed 丧门/披麻/三丘/五墓 shensha annotations"
```

---

## Task 4: 全量验证

- [ ] **Step 1: 全后端测试**

Run: `cd backend && go test ./...`
Expected: 全绿（既有的 `TestGetTokenUsageCostByModel_AggregatesGroupedRows` 在 main 上即 nil *sql.DB panic，与本改动无关，可忽略——其余应 ok）。

- [ ] **Step 2: 编译**

Run: `cd backend && go build ./...`
Expected: 无错误。

- [ ] **Step 3:（可选）人工抽查一个已知盘**

构造或选取一个 `年支=子` 且四柱含 `寅` 的命盘，跑个人报告，确认对应柱出现「丧门」红色凶煞徽章且 tooltip 有释义（migration 应用后）。

---

## 验证标准（对应 spec §6）

1. `go test ./pkg/bazi/...` 全绿（含新增样例）。 → Task 1/2 Step 7/5
2. `go build ./...` 通过。 → Task 4 Step 2
3. 个人报告命中盘显示 4 神煞、红色凶煞徽章、有释义。 → 极性(Task1) + migration(Task3) + 抽查(Task4 Step3)
4. 大运/流年命中运柱出现对应神煞。 → Task 2
5. 全干支严格性：地支对天干错的盘不出现三丘/五墓。 → Task1 `TestPillars_SanQiu_StrictGanZhi`、Task2 `TestDayun_SanQiu_StrictGanZhi`
6. 五墓月柱自命中被排除。 → Task1 `TestPillars_WuMu_ExcludeMonthSelfMatch`
