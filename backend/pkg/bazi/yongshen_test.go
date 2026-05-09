package bazi

import (
	"reflect"
	"strings"
	"testing"
)

// ── 工具函数测试 ────────────────────────────────────────────

func TestCollectNatalGans(t *testing.T) {
	got := collectNatalGans("甲", "丙", "戊", "庚")
	want := []string{"甲", "丙", "戊", "庚"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("collectNatalGans = %v, want %v", got, want)
	}
}

func TestCollectNatalHideGans(t *testing.T) {
	got := collectNatalHideGans(
		[]string{"甲", "丙", "戊"}, // 寅
		[]string{"癸"},            // 子
		[]string{"丁", "己"},       // 午
		[]string{"庚"},            // 申
	)
	if len(got) != 7 {
		t.Errorf("expected 7 hide gans, got %d (%v)", len(got), got)
	}
}

func TestIntersectGans(t *testing.T) {
	hit, miss := intersectGans([]string{"丙", "癸"}, []string{"甲", "丙", "戊"})
	if !reflect.DeepEqual(hit, []string{"丙"}) {
		t.Errorf("hit = %v, want [丙]", hit)
	}
	if !reflect.DeepEqual(miss, []string{"癸"}) {
		t.Errorf("miss = %v, want [癸]", miss)
	}
}

func TestIntersectGansDedup(t *testing.T) {
	// needs 内重复天干应去重
	hit, miss := intersectGans([]string{"丙", "丙", "癸"}, []string{"丙"})
	if !reflect.DeepEqual(hit, []string{"丙"}) {
		t.Errorf("hit dedup = %v, want [丙]", hit)
	}
	if !reflect.DeepEqual(miss, []string{"癸"}) {
		t.Errorf("miss = %v, want [癸]", miss)
	}
}

func TestGansToWuxingSet(t *testing.T) {
	if got := gansToWuxingSet([]string{"丙", "癸"}); got != "火水" {
		t.Errorf("gansToWuxingSet([丙,癸]) = %q, want 火水", got)
	}
	// 6.6：调候要求 ["丁","丙"] 都命中 → yongshen="火"（去重）
	if got := gansToWuxingSet([]string{"丁", "丙"}); got != "火" {
		t.Errorf("gansToWuxingSet([丁,丙]) = %q, want 火（去重）", got)
	}
}

func TestWuxingSetToJishen(t *testing.T) {
	// yongshen=火 → jishen 应含 水(克火) + 土(泄火)
	got := wuxingSetToJishen("火")
	if !strings.Contains(got, "水") || !strings.Contains(got, "土") {
		t.Errorf("wuxingSetToJishen(火) = %q, expected to contain 水 and 土", got)
	}
	// 不应含火本身
	if strings.Contains(got, "火") {
		t.Errorf("wuxingSetToJishen(火) = %q 不应包含火本身", got)
	}
}

// ── 主流程测试 ──────────────────────────────────────────────

// 6.2：t0 命中（透干）：甲日寅月含丙
func TestInferYongshen_T0HitTransparent(t *testing.T) {
	gans := []string{"甲", "丙", "甲", "丁"}                       // 透干含丙
	hideGans := []string{"甲", "丙", "戊", "癸", "甲", "丙", "戊", "丁"} // 藏干（寅+子+寅+巳）
	stats := WuxingStats{Mu: 4, Huo: 2, Tu: 0, Jin: 0, Shui: 1, Total: 8}

	ys, _, status, hit, miss := inferYongshenWithTiaohouPriority(
		"甲", "寅", gans, hideGans, "mu", "mu", stats,
	)
	if status != YongshenStatusTiaohouHit {
		t.Errorf("status = %q, want %q", status, YongshenStatusTiaohouHit)
	}
	if !containsStrTest(hit, "丙") {
		t.Errorf("hit = %v, expected to contain 丙", hit)
	}
	if !strings.Contains(ys, "火") {
		t.Errorf("yongshen = %q, expected to contain 火", ys)
	}
	_ = miss
}

// 6.3：t0 命中（藏干）：甲日寅月四干无丙、月支寅藏丙
func TestInferYongshen_T0HitHidden(t *testing.T) {
	gans := []string{"甲", "甲", "甲", "甲"}                          // 透干无丙
	hideGans := []string{"甲", "丙", "戊", "甲", "丙", "戊"}           // 寅藏含丙
	stats := WuxingStats{Mu: 8, Huo: 0, Tu: 0, Jin: 0, Shui: 0, Total: 8}

	_, _, status, hit, _ := inferYongshenWithTiaohouPriority(
		"甲", "寅", gans, hideGans, "mu", "mu", stats,
	)
	if status != YongshenStatusTiaohouHit {
		t.Errorf("status = %q, want %q (藏干命中也应判 hit)", status, YongshenStatusTiaohouHit)
	}
	if !containsStrTest(hit, "丙") {
		t.Errorf("hit = %v, expected to contain 丙 (from 寅藏)", hit)
	}
}

// 6.4：t0 部分命中：甲日寅月含丙不含癸 → hit=[丙], miss=[癸]
func TestInferYongshen_T0PartialHit(t *testing.T) {
	gans := []string{"甲", "丙", "甲", "丁"}
	hideGans := []string{"甲", "丙", "戊", "丁", "己"} // 没有癸
	stats := WuxingStats{Mu: 4, Huo: 3, Tu: 1, Total: 8}

	_, _, status, hit, miss := inferYongshenWithTiaohouPriority(
		"甲", "寅", gans, hideGans, "mu", "mu", stats,
	)
	if status != YongshenStatusTiaohouHit {
		t.Errorf("status = %q, want %q", status, YongshenStatusTiaohouHit)
	}
	if !reflect.DeepEqual(hit, []string{"丙"}) {
		t.Errorf("hit = %v, want [丙]", hit)
	}
	if !reflect.DeepEqual(miss, []string{"癸"}) {
		t.Errorf("miss = %v, want [癸]", miss)
	}
}

// 6.5：t0 缺位 fallback：甲日寅月透+藏均无丙癸 → fallback to fuyi
func TestInferYongshen_T0Miss(t *testing.T) {
	gans := []string{"甲", "甲", "甲", "甲"}
	hideGans := []string{"甲", "戊", "庚"} // 故意只放无丙无癸
	stats := WuxingStats{Mu: 5, Huo: 0, Tu: 1, Jin: 1, Shui: 1, Total: 8}

	_, _, status, hit, miss := inferYongshenWithTiaohouPriority(
		"甲", "寅", gans, hideGans, "mu", "mu", stats,
	)
	if status != YongshenStatusTiaohouMissFallback {
		t.Errorf("status = %q, want %q", status, YongshenStatusTiaohouMissFallback)
	}
	if len(hit) != 0 {
		t.Errorf("hit = %v, want empty (fallback)", hit)
	}
	// missGans 应记录字典要求但缺位的天干（甲_寅 字典 = 丙癸）
	wantMiss := map[string]bool{"丙": true, "癸": true}
	for _, g := range miss {
		if !wantMiss[g] {
			t.Errorf("miss 含意外天干 %q", g)
		}
	}
}

// 6.7：三冬场景：甲日子月、原局完全无火 → 走 fallback，不再硬编码"火木"
// 子月甲木字典：["丙","庚"]
func TestInferYongshen_ColdMonth_NoFire_FallbackNotHardcoded(t *testing.T) {
	gans := []string{"甲", "壬", "甲", "癸"} // 全无丙庚
	hideGans := []string{"癸", "甲", "癸"}   // 子藏癸
	stats := WuxingStats{Mu: 2, Huo: 0, Tu: 0, Jin: 0, Shui: 4, Total: 8}

	ys, _, status, _, miss := inferYongshenWithTiaohouPriority(
		"甲", "子", gans, hideGans, "mu", "shui", stats,
	)
	if status != YongshenStatusTiaohouMissFallback {
		t.Errorf("status = %q, want fallback (子月调候用神丙庚不在原局)", status)
	}
	// 不应硬编码返回"火木"
	if ys == "火木" {
		t.Errorf("yongshen = %q, 不应再走硬编码'火木'短路", ys)
	}
	if !containsStrTest(miss, "丙") {
		t.Errorf("miss = %v, expected to contain 丙", miss)
	}
}

// 6.8：三夏场景：甲日午月、藏干含癸 → 调候命中"癸"
// 午月甲木字典：["癸","庚","丁"]
func TestInferYongshen_HotMonth_HiddenWaterHit(t *testing.T) {
	gans := []string{"甲", "丙", "甲", "丁"}                  // 透干无癸
	hideGans := []string{"甲", "丙", "戊", "丁", "己", "癸"}   // 藏干含癸（如来自子时藏干）
	stats := WuxingStats{Mu: 3, Huo: 3, Tu: 1, Shui: 1, Total: 8}

	ys, _, status, hit, _ := inferYongshenWithTiaohouPriority(
		"甲", "午", gans, hideGans, "mu", "huo", stats,
	)
	if status != YongshenStatusTiaohouHit {
		t.Errorf("status = %q, want hit (癸 在藏干)", status)
	}
	if !containsStrTest(hit, "癸") {
		t.Errorf("hit = %v, expected to contain 癸", hit)
	}
	if !strings.Contains(ys, "水") {
		t.Errorf("yongshen = %q, expected to contain 水", ys)
	}
}

// 6.9：strings.Contains 兼容性：构造 yongshen=火 的命盘，getYongshenBaseline 仍命中
func TestGetYongshenBaseline_SingleWuxingMatch(t *testing.T) {
	natal := &BaziResult{
		DayGan:   "甲",
		Yongshen: "火",
		Jishen:   "水土",
	}
	pol, ev := getYongshenBaseline(natal, "丙") // 丙=火 → 应命中用神
	if pol != PolarityJi {
		t.Errorf("polarity = %q, want %q (丙=火 应属用神)", pol, PolarityJi)
	}
	if ev == "" {
		t.Errorf("evidence 不应为空")
	}
	// 反向：壬=水 应命中忌神
	pol2, _ := getYongshenBaseline(natal, "壬")
	if pol2 != PolarityXiong {
		t.Errorf("polarity for 壬 = %q, want %q (壬=水 属忌神)", pol2, PolarityXiong)
	}
}

func containsStrTest(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
