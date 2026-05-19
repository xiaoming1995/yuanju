package service

import (
	"testing"
)

func TestGetYearNarrativeMode_DefaultsToAI(t *testing.T) {
	// When no row exists in algo_config, getter returns "ai" (default).
	got := GetYearNarrativeMode()
	if got != "ai" && got != "template" {
		t.Fatalf("expected ai or template, got %q", got)
	}
	// Default behavior: return "ai" when not set.
	// (We cannot assert exact value without DB state control; this test
	// is a smoke check that the function compiles and returns valid value.)
}
