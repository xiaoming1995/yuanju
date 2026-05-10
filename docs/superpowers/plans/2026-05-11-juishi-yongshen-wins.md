# 三合/三会局势力判断——用神赢信号 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增 `collectJuShiSignals` 函数，当流年地支与大运+原局地支补全三合/三会局时，判断局五行与用神/忌神关系，产生"用神赢（吉）"或"忌神局（极凶，标星）"信号，归入 Layer 0 参与压制逻辑。

**Architecture:** 新增独立函数 `collectJuShiSignals`，在 `GetYearEventSignals` 中与现有 `collectYingqiSignals` 并列调用，两者结果合并为 Layer 0 信号集。新增 Type 常量 `TypeJuShiZhong = "局势_重"` 标识极凶忌神局信号。全程 TDD：先写失败测试，再实现最小代码使测试通过。

**Tech Stack:** Go 1.21+，`backend/pkg/bazi` 包，`go test ./pkg/bazi/...`

---

## Files

| 操作 | 路径 | 说明 |
|------|------|------|
| Modify | `backend/pkg/bazi/event_signals.go` | 新增常量、新增函数、修改 Layer 0 收集块 |
| Modify | `backend/pkg/bazi/event_signals_test.go` | 新增 4 个测试 |

---

### Task 1：新增 Type 常量 + 写 4 个失败测试

**Files:**
- Modify: `backend/pkg/bazi/event_signals.go`（常量区，约第 44 行 `EventSignal` struct 前）
- Modify: `backend/pkg/bazi/event_signals_test.go`（末尾追加）

- [ ] **Step 1.1：在 event_signals.go 常量区新增 TypeJuShiZhong**

在文件中找到现有 Type 相关常量（`TypeXueYeZiYuan` 等，约第 34 行），在其后追加：

```go
// TypeJuShiZhong 三合/三会忌神局极凶标星信号
const TypeJuShiZhong = "局势_重"
```

- [ ] **Step 1.2：在 event_signals_test.go 末尾追加 4 个测试**

```go
// ─── 三合/三会局势力判断 ─────────────────────────────────────────────────────

// TestJuShi_YongWins: 用神=火，原局寅+戌，流年=午 → 寅午戌三合火局；忌神=金（火克金）→ 吉
func TestJuShi_YongWins(t *testing.T) {
	natal := makeNatal("壬寅", "甲戌", "壬子", "甲申", "火", "金")
	sigs := collectJuShiSignals(natal, "午", "")
	hasJi := false
	for _, s := range sigs {
		if s.Polarity == PolarityJi && strings.Contains(s.Evidence, "用神赢") {
			hasJi = true
		}
	}
	if !hasJi {
		t.Logf("sigs: %v", sigs)
		t.Fatal("expected 吉 signal with 用神赢 for 三合火局 克 忌神金")
	}
}

// TestJuShi_JiXiong: 忌神=水，原局申+辰，流年=子 → 申子辰三合水局 → 极凶★
func TestJuShi_JiXiong(t *testing.T) {
	natal := makeNatal("壬申", "甲辰", "壬寅", "甲午", "火", "水")
	sigs := collectJuShiSignals(natal, "子", "")
	hasXiong := false
	for _, s := range sigs {
		if s.Type == TypeJuShiZhong && s.Polarity == PolarityXiong && strings.Contains(s.Evidence, "★") {
			hasXiong = true
		}
	}
	if !hasXiong {
		t.Logf("sigs: %v", sigs)
		t.Fatal("expected 极凶 signal Type=局势_重 with ★ for 三合水局（忌神）")
	}
}

// TestJuShi_HalfHe_NoSignal: 原局只有寅，流年=午（缺戌）→ 半合，不触发
func TestJuShi_HalfHe_NoSignal(t *testing.T) {
	natal := makeNatal("壬寅", "甲子", "壬申", "甲辰", "火", "金")
	sigs := collectJuShiSignals(natal, "午", "")
	for _, s := range sigs {
		if strings.Contains(s.Evidence, "用神赢") || strings.Contains(s.Evidence, "寅午戌") {
			t.Fatalf("expected no 三合火局 signal when only half-he, got: %s", s.Evidence)
		}
	}
}

// TestJuShi_NoKe_NoSignal: 用神=火，三合火局成，忌神=土（火不克土）→ 无信号
func TestJuShi_NoKe_NoSignal(t *testing.T) {
	natal := makeNatal("壬寅", "甲戌", "壬子", "甲申", "火", "土")
	sigs := collectJuShiSignals(natal, "午", "")
	for _, s := range sigs {
		if strings.Contains(s.Evidence, "用神赢") {
			t.Fatalf("expected no 用神赢 signal when 火 does not ke 土, got: %s", s.Evidence)
		}
	}
}
```

- [ ] **Step 1.3：确认测试失败（函数未定义）**

```bash
cd backend && go test ./pkg/bazi/... -run "TestJuShi" 2>&1
```

期望输出包含：`undefined: collectJuShiSignals`

---

### Task 2：实现 collectJuShiSignals

**Files:**
- Modify: `backend/pkg/bazi/event_signals.go`（在 `collectYingqiSignals` 函数之前插入新函数）

- [ ] **Step 2.1：在 event_signals.go 中插入 collectJuShiSignals**

在 `pillarWeightLabel` 函数定义之后、`collectYingqiSignals` 之前，插入：

```go
// juGroup 三合/三会局组定义
type juGroup struct {
	branches [3]string
	wx       string // 局五行（pinyin）
	kind     string // "三合" 或 "三会"
}

// allJuGroups 所有三合/三会局
var allJuGroups = []juGroup{
	// 三合
	{[3]string{"申", "子", "辰"}, "shui", "三合"},
	{[3]string{"寅", "午", "戌"}, "huo", "三合"},
	{[3]string{"亥", "卯", "未"}, "mu", "三合"},
	{[3]string{"巳", "酉", "丑"}, "jin", "三合"},
	// 三会
	{[3]string{"寅", "卯", "辰"}, "mu", "三会"},
	{[3]string{"巳", "午", "未"}, "huo", "三会"},
	{[3]string{"申", "酉", "戌"}, "jin", "三会"},
	{[3]string{"亥", "子", "丑"}, "shui", "三会"},
}

// collectJuShiSignals 检测流年地支（结合大运+原局）是否补全三合/三会局。
// 三支全齐（matchCount=2）才算局成；局五行=用神且克忌神→吉；局五行=忌神→极凶（TypeJuShiZhong）。
func collectJuShiSignals(natal *BaziResult, lnZhi, dyZhi string) []EventSignal {
	if natal == nil || lnZhi == "" {
		return nil
	}
	if natal.Yongshen == "" && natal.Jishen == "" {
		return nil
	}

	existingZhi := []string{natal.YearZhi, natal.MonthZhi, natal.DayZhi, natal.HourZhi}
	if dyZhi != "" {
		existingZhi = append(existingZhi, dyZhi)
	}

	cn2pin := map[string]string{"木": "mu", "火": "huo", "土": "tu", "金": "jin", "水": "shui"}
	yongSet := map[string]bool{}
	jiSet := map[string]bool{}
	for _, ch := range natal.Yongshen {
		if p, ok := cn2pin[string(ch)]; ok {
			yongSet[p] = true
		}
	}
	for _, ch := range natal.Jishen {
		if p, ok := cn2pin[string(ch)]; ok {
			jiSet[p] = true
		}
	}

	var sigs []EventSignal
	for _, g := range allJuGroups {
		// 确认 lnZhi 在该组中
		lnIdx := -1
		for i, z := range g.branches {
			if z == lnZhi {
				lnIdx = i
				break
			}
		}
		if lnIdx < 0 {
			continue
		}
		// 另外两支必须都在 existingZhi 中（三支全齐）
		other := make([]string, 0, 2)
		for i, z := range g.branches {
			if i != lnIdx {
				other = append(other, z)
			}
		}
		if !containsStr(existingZhi, other[0]) || !containsStr(existingZhi, other[1]) {
			continue
		}

		localWx := g.wx
		juName := string(g.branches[0]) + string(g.branches[1]) + string(g.branches[2])
		localWxCN := wxPinyin2CN[localWx]

		if yongSet[localWx] {
			keTarget := wxKe[localWx]
			if jiSet[keTarget] {
				jiCN := wxPinyin2CN[keTarget]
				sigs = append(sigs, EventSignal{
					Type:     "综合变动",
					Evidence: fmt.Sprintf("流年%s补全%s%s%s局，%s势力大增，克制忌神%s，用神赢，应期吉", lnZhi, juName, g.kind, localWxCN, localWxCN, jiCN),
					Polarity: PolarityJi,
					Source:   SourceZhuwei,
				})
			}
		} else if jiSet[localWx] {
			sigs = append(sigs, EventSignal{
				Type:     TypeJuShiZhong,
				Evidence: fmt.Sprintf("★流年%s补全%s%s%s局，忌神势力极强，用神承压，应期极凶", lnZhi, juName, g.kind, localWxCN),
				Polarity: PolarityXiong,
				Source:   SourceZhuwei,
			})
		}
	}
	return sigs
}
```

- [ ] **Step 2.2：运行 4 个新测试**

```bash
cd backend && go test ./pkg/bazi/... -run "TestJuShi" -v 2>&1
```

期望：4 个测试全部 PASS。

- [ ] **Step 2.3：运行全量测试确认无回归**

```bash
cd backend && go test ./pkg/bazi/... 2>&1
```

期望：`ok  yuanju/pkg/bazi`

- [ ] **Step 2.4：Commit**

```bash
git add backend/pkg/bazi/event_signals.go backend/pkg/bazi/event_signals_test.go
git commit -m "feat(bazi): 新增三合/三会局势力判断——用神赢/忌神局信号"
```

---

### Task 3：集成到 GetYearEventSignals Layer 0

**Files:**
- Modify: `backend/pkg/bazi/event_signals.go`（`GetYearEventSignals` 函数内，约 1176 行）

- [ ] **Step 3.1：修改 Layer 0 收集块**

找到以下代码块（约 1176 行）：

```go
	// ── 应期位置信号（Layer 0：刑冲克合穿破原局用神/忌神位）────────────────────
	layer0Sigs := collectYingqiSignals(natal, lnGan, lnZhi, dyGan, dyZhi)
	layer0HasXiong := false
	for _, s := range layer0Sigs {
		if s.Polarity == PolarityXiong {
			layer0HasXiong = true
			break
		}
	}
	signals = append(signals, layer0Sigs...)
	layer0End := len(signals) // Layer 0 ends here; Layer 4+ signals begin after
```

替换为：

```go
	// ── 应期位置信号（Layer 0：刑冲克合穿破原局用神/忌神位 + 三合/三会局势力）──
	layer0Sigs := collectYingqiSignals(natal, lnGan, lnZhi, dyGan, dyZhi)
	layer0Sigs = append(layer0Sigs, collectJuShiSignals(natal, lnZhi, dyZhi)...)
	layer0HasXiong := false
	for _, s := range layer0Sigs {
		if s.Polarity == PolarityXiong {
			layer0HasXiong = true
			break
		}
	}
	signals = append(signals, layer0Sigs...)
	layer0End := len(signals) // Layer 0 ends here; Layer 4+ signals begin after
```

- [ ] **Step 3.2：运行全量测试**

```bash
cd backend && go test ./pkg/bazi/... 2>&1
```

期望：`ok  yuanju/pkg/bazi`

- [ ] **Step 3.3：Commit**

```bash
git add backend/pkg/bazi/event_signals.go
git commit -m "feat(bazi): 将三合/三会局势力信号纳入 Layer 0 压制逻辑"
```

---

## 自检结果

- **Spec 覆盖**：三合/三会全齐判断 ✓、用神赢（吉）✓、忌神局（极凶+★+新 Type）✓、静默跳过（无克/半合）✓、Layer 0 集成 ✓
- **Placeholder**：无 TBD/TODO
- **类型一致性**：`TypeJuShiZhong`、`collectJuShiSignals`、`allJuGroups`、`juGroup` 在所有 Task 中命名一致
