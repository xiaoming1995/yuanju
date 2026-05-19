package service

import (
	"strings"
	"testing"

	"yuanju/pkg/bazi"
)

func TestValidateYearNarrative_PassesWhenAllKeywordsTraceToEvidence(t *testing.T) {
	signals := []bazi.EventSignal{
		{Type: "健康", Evidence: "白虎临运，主孝服、突发伤痛或意外", Source: bazi.SourceShensha, Polarity: bazi.PolarityXiong},
		{Type: "迁变", Evidence: "驿马临运，主奔波、出行、变动", Source: bazi.SourceShensha, Polarity: bazi.PolarityNeutral},
	}
	narrative := "本年白虎临运，健康注意；驿马合年支，宜防奔波。"
	ok, reason := ValidateYearNarrative(narrative, signals)
	if !ok {
		t.Errorf("expected pass, got fail: %s", reason)
	}
}

func TestValidateYearNarrative_FailsOnUnattestedShensha(t *testing.T) {
	signals := []bazi.EventSignal{
		{Type: "迁变", Evidence: "驿马临运，主奔波", Source: bazi.SourceShensha, Polarity: bazi.PolarityNeutral},
	}
	// AI fabricated "桃花临运" which doesn't appear in any signal evidence.
	narrative := "本年驿马动象明显，桃花临运人缘旺。"
	ok, reason := ValidateYearNarrative(narrative, signals)
	if ok {
		t.Errorf("expected fail (桃花 not in evidence), got pass")
	}
	if !strings.Contains(reason, "桃花") {
		t.Errorf("expected reason to mention 桃花, got: %s", reason)
	}
}

func TestValidateYearNarrative_FailsOnUnattestedPositionClaim(t *testing.T) {
	signals := []bazi.EventSignal{
		{Type: "综合变动", Evidence: "流年地支辰冲月柱戌", Source: bazi.SourceZhuwei, Polarity: bazi.PolarityXiong},
	}
	// AI fabricated "用神位受刑" — no signal has that text.
	narrative := "流年地支辰冲月柱戌，且用神位受刑，宜防。"
	ok, reason := ValidateYearNarrative(narrative, signals)
	if ok {
		t.Errorf("expected fail (用神位 not attested), got pass")
	}
	if !strings.Contains(reason, "用神位") {
		t.Errorf("expected reason to mention 用神位, got: %s", reason)
	}
}

func TestValidateYearNarrative_PassesWithoutAnyHardKeywords(t *testing.T) {
	// Narrative with no validated keywords passes trivially.
	signals := []bazi.EventSignal{
		{Type: "事业", Evidence: "流年节奏微调", Source: bazi.SourceZhuwei, Polarity: bazi.PolarityNeutral},
	}
	narrative := "本年事业上节奏微调，按部就班即可。"
	ok, _ := ValidateYearNarrative(narrative, signals)
	if !ok {
		t.Error("expected pass for narrative without validated keywords")
	}
}

func TestValidateYearNarrative_PassesOnEmptyNarrative(t *testing.T) {
	// Empty narrative (AI explicitly skipped this year) always passes.
	signals := []bazi.EventSignal{}
	ok, _ := ValidateYearNarrative("", signals)
	if !ok {
		t.Error("expected pass for empty narrative")
	}
}

func TestValidateYearNarrative_FuyinFanyinDayunheHua(t *testing.T) {
	// 伏吟 / 反吟 / 大运合化 are validated terms.
	signals := []bazi.EventSignal{
		{Type: "伏吟", Evidence: "流年壬辰伏吟日柱壬辰，主同类事件重现", Source: bazi.SourceFuyin, Polarity: bazi.PolarityXiong},
	}
	narrative := "本年伏吟日柱，旧事重提；反吟未现。"
	ok, reason := ValidateYearNarrative(narrative, signals)
	if ok {
		t.Errorf("expected fail (反吟 not in evidence), got pass")
	}
	if !strings.Contains(reason, "反吟") {
		t.Errorf("expected reason to mention 反吟, got: %s", reason)
	}
}

// ── ExtractEvidenceKeywords 测试 ───────────────────────────────────────

func TestExtractEvidenceKeywords_SingleNeshalSignal(t *testing.T) {
	signals := []bazi.EventSignal{
		{Type: "迁变", Evidence: "驿马临运，主奔波、出行、变动"},
	}
	out := ExtractEvidenceKeywords(signals)
	if len(out) != 1 || out[0] != "驿马" {
		t.Errorf("expected [驿马], got %v", out)
	}
}

func TestExtractEvidenceKeywords_MultipleDistinctKeywords(t *testing.T) {
	signals := []bazi.EventSignal{
		{Type: "综合变动", Evidence: "流年甲冲原局月柱戊（用神位），用神受冲"},
		{Type: "健康", Evidence: "羊刃临运，宜防开刀、血光、车祸"},
		{Type: "迁变", Evidence: "驿马临运，主奔波"},
	}
	out := ExtractEvidenceKeywords(signals)
	// 验证至少包含这些核心关键词
	required := []string{"用神位", "受冲", "羊刃", "驿马"}
	for _, kw := range required {
		found := false
		for _, k := range out {
			if k == kw {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q in keywords, got %v", kw, out)
		}
	}
}

func TestExtractEvidenceKeywords_DeduplicatesAcrossSignals(t *testing.T) {
	signals := []bazi.EventSignal{
		{Type: "综合变动", Evidence: "流年与日柱伏吟"},
		{Type: "伏吟", Evidence: "干支双伏吟，应期极强"},
	}
	out := ExtractEvidenceKeywords(signals)
	count := 0
	for _, k := range out {
		if k == "伏吟" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected dedup to 1 occurrence of 伏吟, got %d", count)
	}
}

func TestExtractEvidenceKeywords_EmptySignalsReturnsNil(t *testing.T) {
	if out := ExtractEvidenceKeywords(nil); out != nil {
		t.Errorf("expected nil for nil signals, got %v", out)
	}
	if out := ExtractEvidenceKeywords([]bazi.EventSignal{}); out != nil {
		t.Errorf("expected nil for empty signals, got %v", out)
	}
}

func TestExtractEvidenceKeywords_NoKnownKeywordReturnsEmpty(t *testing.T) {
	signals := []bazi.EventSignal{
		{Type: "综合变动", Evidence: "流年事件，无神煞触发"},
	}
	out := ExtractEvidenceKeywords(signals)
	if len(out) != 0 {
		t.Errorf("expected no keywords, got %v", out)
	}
}
