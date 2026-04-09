package bazi

import (
	"testing"
)

// ─── Bug 1a: 魁罡 仅限日柱 ───────────────────────────────────────────────────

// 年柱=庚辰（魁罡干支），日柱=甲午（非魁罡）→ 魁罡不应出现在年柱
func TestKuiGangOnlyInDayPillar(t *testing.T) {
	result := GetPillarsShenSha("庚", "辰", "丙", "子", "甲", "午", "壬", "子")
	for _, s := range result[0] {
		if s == "魁罡" {
			t.Error("魁罡出现在年柱，应仅限日柱")
		}
	}
	for _, s := range result[1] {
		if s == "魁罡" {
			t.Error("魁罡出现在月柱，应仅限日柱")
		}
	}
}

// 日柱=庚辰（魁罡干支）→ 日柱应有魁罡（正向验证）
func TestKuiGangInDayPillar(t *testing.T) {
	result := GetPillarsShenSha("甲", "子", "丙", "子", "庚", "辰", "壬", "子")
	found := false
	for _, s := range result[2] {
		if s == "魁罡" {
			found = true
		}
	}
	if !found {
		t.Error("日柱=庚辰，应有魁罡")
	}
}

// ─── Bug 1b: 十恶大败 仅限日柱 ───────────────────────────────────────────────

// 年柱=甲辰（十恶大败干支），日柱=甲午（非十恶大败）→ 十恶大败不应在年柱
func TestShiEOnlyInDayPillar(t *testing.T) {
	result := GetPillarsShenSha("甲", "辰", "丙", "子", "甲", "午", "壬", "子")
	for _, s := range result[0] {
		if s == "十恶大败" {
			t.Error("十恶大败出现在年柱，应仅限日柱")
		}
	}
}

// 日柱=甲辰（十恶大败干支）→ 日柱应有十恶大败（正向验证）
func TestShiEInDayPillar(t *testing.T) {
	result := GetPillarsShenSha("壬", "子", "丙", "子", "甲", "辰", "壬", "子")
	found := false
	for _, s := range result[2] {
		if s == "十恶大败" {
			found = true
		}
	}
	if !found {
		t.Error("日柱=甲辰，应有十恶大败")
	}
}

// ─── Bug 1c: 日德 仅限日柱 + 定义修正 ────────────────────────────────────────

// 日柱=甲辰（旧错误定义 + 同时是十恶大败）→ 不应同时有日德和十恶大败
func TestRiDeNotCoexistWithShiE(t *testing.T) {
	result := GetPillarsShenSha("壬", "子", "丙", "子", "甲", "辰", "壬", "子")
	hasRiDe, hasShiE := false, false
	for _, s := range result[2] {
		if s == "日德" {
			hasRiDe = true
		}
		if s == "十恶大败" {
			hasShiE = true
		}
	}
	if hasRiDe && hasShiE {
		t.Error("同一柱不应同时出现日德（吉）和十恶大败（凶）")
	}
}

// 日柱=甲寅（新日德定义）→ 日柱应有日德
func TestRiDeJiaYin(t *testing.T) {
	result := GetPillarsShenSha("壬", "子", "丙", "子", "甲", "寅", "壬", "子")
	found := false
	for _, s := range result[2] {
		if s == "日德" {
			found = true
		}
	}
	if !found {
		t.Error("日柱=甲寅，应有日德")
	}
}

// 日柱=丙午（新日德定义）→ 日柱应有日德
func TestRiDeBingWu(t *testing.T) {
	result := GetPillarsShenSha("壬", "子", "丙", "子", "丙", "午", "壬", "子")
	found := false
	for _, s := range result[2] {
		if s == "日德" {
			found = true
		}
	}
	if !found {
		t.Error("日柱=丙午，应有日德")
	}
}

// 年柱=甲寅（日德干支），但日德应仅限日柱 → 年柱不应有日德
func TestRiDeOnlyInDayPillar(t *testing.T) {
	result := GetPillarsShenSha("甲", "寅", "丙", "子", "丙", "午", "壬", "子")
	for _, s := range result[0] {
		if s == "日德" {
			t.Error("日德出现在年柱，应仅限日柱")
		}
	}
}

// ─── Bug 5: 天喜 改用红鸾天喜口诀 ───────────────────────────────────────────

// 亥年天喜=戌（红鸾天喜表），月支=戌 → 月柱应有天喜
func TestTianxiHaiYear(t *testing.T) {
	result := GetPillarsShenSha("乙", "亥", "丙", "戌", "丙", "子", "甲", "午")
	found := false
	for _, s := range result[1] {
		if s == "天喜" {
			found = true
		}
	}
	if !found {
		t.Error("亥年天喜=戌，月支=戌，月柱应有天喜")
	}
}

// 子年天喜=酉 → 正向验证另一年份
func TestTianxiZiYear(t *testing.T) {
	// 子年(如2008年戊子)，天喜=酉
	result := GetPillarsShenSha("戊", "子", "甲", "子", "甲", "酉", "壬", "子")
	found := false
	for _, s := range result[2] {
		if s == "天喜" {
			found = true
		}
	}
	if !found {
		t.Error("子年天喜=酉，日支=酉，日柱应有天喜")
	}
}

// ─── Bug 6: 德秀贵人 ─────────────────────────────────────────────────────────

// 戌月天德=丙，月德=丙（相同），月柱干=丙 → 月柱应有德秀贵人
func TestDeshuiXuMonth(t *testing.T) {
	result := GetPillarsShenSha("乙", "亥", "丙", "戌", "丙", "子", "甲", "午")
	found := false
	for _, s := range result[1] {
		if s == "德秀贵人" {
			found = true
		}
	}
	if !found {
		t.Error("戌月天德=月德=丙，月柱干=丙，月柱应有德秀贵人")
	}
}

// 寅月天德=丁，月德=丙（不同），丙干柱不应有德秀贵人
func TestDeshuiNotWhenDiff(t *testing.T) {
	// 寅月: 天德=丁, 月德=丙, 不同 → 无德秀贵人
	result := GetPillarsShenSha("甲", "子", "丙", "寅", "丙", "午", "甲", "子")
	for _, s := range result[2] {
		if s == "德秀贵人" {
			t.Error("寅月天德=丁≠月德=丙，不应有德秀贵人")
		}
	}
}

// ─── Bug 3: 国印贵人 改为各柱自查 ────────────────────────────────────────────

// 年柱=乙亥：乙干对应国印地支=亥，年支=亥 → 年柱应有国印贵人
func TestGuoYinSelfCheck(t *testing.T) {
	result := GetPillarsShenSha("乙", "亥", "丙", "子", "丙", "子", "甲", "午")
	found := false
	for _, s := range result[0] {
		if s == "国印贵人" {
			found = true
		}
	}
	if !found {
		t.Error("年柱乙亥：乙干国印地支=亥，年支=亥，年柱应有国印贵人")
	}
}

// 日柱=丙子：丙→丑，日支=子≠丑 → 日柱不应有国印贵人
func TestGuoYinDayNone(t *testing.T) {
	result := GetPillarsShenSha("乙", "亥", "丙", "子", "丙", "子", "甲", "午")
	for _, s := range result[2] {
		if s == "国印贵人" {
			t.Error("日柱丙子：丙→丑，日支=子，不应有国印贵人")
		}
	}
}

// ─── Bug 4: 天罗只标戌，地网只标辰 ──────────────────────────────────────────

// 亥戌同现 → 天罗只标月柱(戌)，不标年柱(亥)
func TestTianluoOnlyOnXu(t *testing.T) {
	result := GetPillarsShenSha("乙", "亥", "丙", "戌", "丙", "子", "甲", "午")
	for _, s := range result[0] {
		if s == "天罗" {
			t.Error("天罗不应出现在亥(年柱)，仅限戌所在柱")
		}
	}
	found := false
	for _, s := range result[1] {
		if s == "天罗" {
			found = true
		}
	}
	if !found {
		t.Error("戌(月柱)应有天罗")
	}
}

// 辰巳同现 → 地网只标年柱(辰)，不标月柱(巳)
func TestDiwangOnlyOnChen(t *testing.T) {
	result := GetPillarsShenSha("甲", "辰", "丙", "巳", "壬", "子", "甲", "午")
	for _, s := range result[1] {
		if s == "地网" {
			t.Error("地网不应出现在巳(月柱)，仅限辰所在柱")
		}
	}
	found := false
	for _, s := range result[0] {
		if s == "地网" {
			found = true
		}
	}
	if !found {
		t.Error("辰(年柱)应有地网")
	}
}

// ─── Bug 2: 卯/午/酉月 天德合（地支型）─────────────────────────────────────

// 卯月：天德=申（地支型），天德合=巳（巳申六合）
// 日柱地支=巳 → 日柱应有天德合
func TestTiandeheMaoMonth(t *testing.T) {
	result := GetPillarsShenSha("甲", "子", "丙", "卯", "甲", "巳", "壬", "子")
	found := false
	for _, s := range result[2] {
		if s == "天德合" {
			found = true
		}
	}
	if !found {
		t.Error("卯月（天德=申），日柱地支=巳（巳申六合），应有天德合")
	}
}

// 午月：天德=亥（地支型），天德合=寅（寅亥六合）
// 日柱地支=寅 → 日柱应有天德合
func TestTiandeheWuMonth(t *testing.T) {
	result := GetPillarsShenSha("甲", "子", "丙", "午", "甲", "寅", "壬", "子")
	found := false
	for _, s := range result[2] {
		if s == "天德合" {
			found = true
		}
	}
	if !found {
		t.Error("午月（天德=亥），日柱地支=寅（寅亥六合），应有天德合")
	}
}

// 酉月：天德=寅（地支型），天德合=亥（寅亥六合）
// 日柱地支=亥 → 日柱应有天德合
func TestTiandeheYouMonth(t *testing.T) {
	result := GetPillarsShenSha("甲", "子", "丙", "酉", "甲", "亥", "壬", "子")
	found := false
	for _, s := range result[2] {
		if s == "天德合" {
			found = true
		}
	}
	if !found {
		t.Error("酉月（天德=寅），日柱地支=亥（寅亥六合），应有天德合")
	}
}
