package handler

import (
	"strings"
	"testing"
)

func TestNormalizeChartDisplayName(t *testing.T) {
	name, err := normalizeChartDisplayName("  小王  ")
	if err != nil {
		t.Fatal(err)
	}
	if name != "小王" {
		t.Fatalf("expected trimmed name, got %q", name)
	}
}

func TestNormalizeChartDisplayName_AllowsEmpty(t *testing.T) {
	name, err := normalizeChartDisplayName("   ")
	if err != nil {
		t.Fatal(err)
	}
	if name != "" {
		t.Fatalf("expected empty name, got %q", name)
	}
}

func TestNormalizeChartDisplayName_RejectsLongName(t *testing.T) {
	_, err := normalizeChartDisplayName(strings.Repeat("命", 21))
	if err == nil {
		t.Fatal("expected long name to be rejected")
	}
	if !strings.Contains(err.Error(), "20") {
		t.Fatalf("expected error to mention length limit, got %v", err)
	}
}
