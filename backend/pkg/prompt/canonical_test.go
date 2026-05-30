package prompt

import (
	"strings"
	"testing"
)

func TestMustGet_CompatibilityReturnsRegisteredDefinition(t *testing.T) {
	def := MustGet("compatibility")
	if def.Content == "" {
		t.Fatal("compatibility canonical content must not be empty")
	}
	if def.Version != "v3.1-question-aware-4" {
		t.Errorf("expected Version v3.1-question-aware-4, got %q", def.Version)
	}
	if len(def.Hash) != 64 {
		t.Errorf("Hash must be 64-char sha256 hex, got len %d: %q", len(def.Hash), def.Hash)
	}
	// Sanity: prompt body should reference the new structured schema
	for _, want := range []string{"question_focus", "decision_advice", "personality_comparison", "表达约束", "{{.PrimaryQuestionLabel}}"} {
		if !strings.Contains(def.Content, want) {
			t.Errorf("compatibility canonical content missing %q", want)
		}
	}
}

func TestCompatibilityPromptUsesV3ModuleKeys(t *testing.T) {
	def, ok := Lookup("compatibility")
	if !ok {
		t.Fatal("compatibility prompt not registered")
	}
	content := def.Content
	// v3 prompt must use 4-module keys, not legacy 4-dim keys
	for _, want := range []string{`"key": "zodiac"`, `"key": "nayin"`, `"key": "day_pillar"`, `"key": "eight_chars"`} {
		if !strings.Contains(content, want) {
			t.Errorf("compatibility prompt missing v3 module key %q", want)
		}
	}
	for _, forbidden := range []string{`"key": "attraction"`, `"key": "stability"`, `"key": "practicality"`} {
		if strings.Contains(content, forbidden) {
			t.Errorf("compatibility prompt still references legacy dim key %q", forbidden)
		}
	}
}

func TestMustGet_UnknownModulePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unknown module")
		}
	}()
	MustGet("not-a-real-module")
}
