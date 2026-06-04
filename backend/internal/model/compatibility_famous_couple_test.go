package model

import (
	"encoding/json"
	"testing"
)

// LLM 返回的 JSON 经 Unmarshal→Marshal 持久化后，famous_couple 必须存活。
func TestCompatibilityStructuredReport_FamousCoupleSurvivesRoundTrip(t *testing.T) {
	raw := `{
		"summary": "x",
		"famous_couple": {
			"couple": "梁山伯与祝英台",
			"tagline": "一见倾心，却被现实层层阻隔",
			"reason": "你们吸引力来得快而强，但长期更受现实安排牵制。"
		},
		"risks": [],
		"advice": "y"
	}`

	var report CompatibilityStructuredReport
	if err := json.Unmarshal([]byte(raw), &report); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if report.FamousCouple == nil {
		t.Fatal("FamousCouple was dropped on unmarshal")
	}
	if report.FamousCouple.Couple != "梁山伯与祝英台" {
		t.Errorf("Couple = %q, want 梁山伯与祝英台", report.FamousCouple.Couple)
	}

	out, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var again CompatibilityStructuredReport
	if err := json.Unmarshal(out, &again); err != nil {
		t.Fatalf("re-unmarshal failed: %v", err)
	}
	if again.FamousCouple == nil || again.FamousCouple.Tagline != "一见倾心，却被现实层层阻隔" {
		t.Errorf("famous_couple lost across round-trip: %+v", again.FamousCouple)
	}
}

// 旧报告没有 famous_couple 时，字段应为 nil 且 marshal 时省略（omitempty）。
func TestCompatibilityStructuredReport_FamousCoupleOmittedWhenAbsent(t *testing.T) {
	var report CompatibilityStructuredReport
	if err := json.Unmarshal([]byte(`{"summary":"x","risks":[],"advice":"y"}`), &report); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if report.FamousCouple != nil {
		t.Fatalf("expected nil FamousCouple, got %+v", report.FamousCouple)
	}
	out, _ := json.Marshal(report)
	if string(out) == "" {
		t.Fatal("marshal produced empty output")
	}
	if containsKey(out, "famous_couple") {
		t.Errorf("famous_couple should be omitted when nil, got: %s", out)
	}
}

func containsKey(b []byte, key string) bool {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return false
	}
	_, ok := m[key]
	return ok
}
