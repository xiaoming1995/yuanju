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
	if def.Version != "v3-question-aware" {
		t.Errorf("expected Version v3-question-aware, got %q", def.Version)
	}
	if len(def.Hash) != 64 {
		t.Errorf("Hash must be 64-char sha256 hex, got len %d: %q", len(def.Hash), def.Hash)
	}
	// Sanity: prompt body should reference the new structured schema
	for _, want := range []string{"question_focus", "decision_advice", "{{.PrimaryQuestionLabel}}"} {
		if !strings.Contains(def.Content, want) {
			t.Errorf("compatibility canonical content missing %q", want)
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
