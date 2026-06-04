package prompt

import (
	"strings"
	"testing"
)

func TestCompatibilityPromptIncludesFamousCouple(t *testing.T) {
	def := MustGet("compatibility")
	// schema 字段
	for _, want := range []string{
		`"famous_couple"`,
		`"couple"`,
		`"tagline"`,
		`"reason"`,
	} {
		if !strings.Contains(def.Content, want) {
			t.Errorf("compatibility prompt missing famous_couple schema token %q", want)
		}
	}
	// 约束段：必须反映真实动态、可悲剧、得体不越线
	for _, want := range []string{
		"名人类比约束",
		"真实动态",
	} {
		if !strings.Contains(def.Content, want) {
			t.Errorf("compatibility prompt missing famous_couple constraint %q", want)
		}
	}
}
