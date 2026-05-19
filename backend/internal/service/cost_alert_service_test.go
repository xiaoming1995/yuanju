package service

import (
	"testing"
)

func TestBuildBudgetStatus_SmokeCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("requires DB")
	}
	status, err := BuildBudgetStatus()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if status.Today.ThresholdCostCny <= 0 {
		t.Errorf("today threshold should be > 0, got %v", status.Today.ThresholdCostCny)
	}
	if status.ThisMonth.ThresholdCostCny <= 0 {
		t.Errorf("month threshold should be > 0, got %v", status.ThisMonth.ThresholdCostCny)
	}
	if status.PerChartThresholdCny <= 0 {
		t.Errorf("per-chart threshold should be > 0, got %v", status.PerChartThresholdCny)
	}
	if status.Today.TotalCostCny > status.Today.ThresholdCostCny && !status.Today.Exceeded {
		t.Error("Today.Exceeded should be true when totalCost > threshold")
	}
}

func TestExceededPct_HandlesZeroThreshold(t *testing.T) {
	pct := exceededPct(50, 0)
	if pct != 0 {
		t.Errorf("expected 0 for zero threshold, got %d", pct)
	}
	pct = exceededPct(50, 100)
	if pct != 50 {
		t.Errorf("expected 50, got %d", pct)
	}
	pct = exceededPct(150, 100)
	if pct != 150 {
		t.Errorf("expected 150, got %d", pct)
	}
}
