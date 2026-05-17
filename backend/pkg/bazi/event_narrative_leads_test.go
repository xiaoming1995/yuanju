package bazi

import (
	"strings"
	"testing"
)

func assertAllDistinct(t *testing.T, name string, outs []string) {
	t.Helper()
	seen := map[string]int{}
	for i, s := range outs {
		if s == "" {
			t.Fatalf("%s output #%d is empty", name, i)
		}
		if prev, ok := seen[s]; ok {
			t.Fatalf("%s output #%d duplicates #%d: %q", name, i, prev, s)
		}
		seen[s] = i
	}
}

func TestChangeLeadDistinctBranches(t *testing.T) {
	inputs := []EventSignal{
		{Type: "伏吟", Polarity: PolarityXiong, Source: SourceFuyin},
		{Type: "反吟", Polarity: PolarityXiong, Source: SourceXing},
		{Type: "大运合化", Polarity: PolarityXiong, Source: SourceHehua},
		{Type: TypeJuShiZhong, Polarity: PolarityXiong},
		{Type: "综合变动", Polarity: PolarityXiong, Source: SourceXing, Evidence: "受刑"},
	}
	outs := make([]string, len(inputs))
	for i, in := range inputs {
		outs[i] = changeLead(in)
	}
	assertAllDistinct(t, "changeLead", outs)
}

func TestChangeLeadLegacyFallbackReachable(t *testing.T) {
	sig := EventSignal{
		Type:     "综合变动",
		Polarity: PolarityXiong,
		Source:   "",
		Evidence: "月柱受冲",
	}
	got := changeLead(sig)
	want := "这一年的变动感比较强，旧问题或突发调整容易被推到眼前"
	if got != want {
		t.Fatalf("changeLead legacy fallback = %q, want %q", got, want)
	}
}

func TestHealthLeadThreeBranches(t *testing.T) {
	inputs := []EventSignal{
		{Type: "健康", Polarity: PolarityXiong, Evidence: "白虎临运"},
		{Type: "健康", Polarity: PolarityXiong, Evidence: "羊刃临运"},
		{Type: "健康", Polarity: PolarityJi, Evidence: "天医临运"},
	}
	outs := make([]string, len(inputs))
	for i, in := range inputs {
		outs[i] = healthLead(in)
	}
	assertAllDistinct(t, "healthLead", outs)
}

func TestRelationshipLeadDistinctBranches(t *testing.T) {
	inputs := []EventSignal{
		{Type: "婚恋_合", Polarity: PolarityJi},
		{Type: "婚恋_冲", Polarity: PolarityXiong},
		{Type: "婚恋_变", Polarity: PolarityXiong},
		{Type: TypeXingGeQingYi, Polarity: PolarityJi},
		{Type: TypeXingGePanNi, Polarity: PolarityXiong},
	}
	outs := make([]string, len(inputs))
	for i, in := range inputs {
		outs[i] = relationshipLead(in)
	}
	assertAllDistinct(t, "relationshipLead", outs)
}

func TestDefaultHardLeadFourSources(t *testing.T) {
	inputs := []EventSignal{
		{Source: SourceKongwang, Polarity: PolarityXiong},
		{Source: SourceXing, Polarity: PolarityXiong},
		{Source: SourceFuyin, Polarity: PolarityXiong},
		{Source: SourceHehua, Polarity: PolarityXiong},
	}
	outs := make([]string, len(inputs))
	for i, in := range inputs {
		outs[i] = defaultHardLead(in)
	}
	assertAllDistinct(t, "defaultHardLead", outs)
}

// 防回归：旧的 "这一年的变动感比较强" 字符串必须仍被 changeLead 兜底分支保留
func TestLegacyChangeLeadStringStillPresent(t *testing.T) {
	src := EventSignal{Type: "综合变动", Polarity: PolarityXiong}
	if !strings.Contains(changeLead(src), "这一年的变动感比较强") {
		t.Fatalf("legacy change lead string must remain in fallback branch")
	}
}
