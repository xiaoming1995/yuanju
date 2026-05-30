package service

import (
	"testing"
	"time"

	"yuanju/pkg/database"
)

func TestBuildBudgetStatus_SmokeCheck(t *testing.T) {
	if testing.Short() || database.DB == nil {
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
	if status.PerUserThresholdCny <= 0 {
		t.Errorf("per-user threshold should be > 0, got %v", status.PerUserThresholdCny)
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

func TestAlertDedup_FirstCrossLogsThenSilences(t *testing.T) {
	state := newAlertState()

	// First exceeded → should log
	if !state.shouldAlert("daily_total", time.Now()) {
		t.Error("first exceeded crossing should log")
	}
	state.markAlerted("daily_total", time.Now())

	// Immediately again → should NOT log (within 1h window)
	if state.shouldAlert("daily_total", time.Now().Add(30*time.Minute)) {
		t.Error("within 1h dedup window should not log")
	}

	// >1h later → should log again
	if !state.shouldAlert("daily_total", time.Now().Add(2*time.Hour)) {
		t.Error("after 1h should log again")
	}
}

func TestAlertState_DifferentScopesIndependent(t *testing.T) {
	state := newAlertState()
	now := time.Now()
	state.markAlerted("daily_total", now)

	// 不同 scope 的告警互不影响
	if !state.shouldAlert("monthly_total", now) {
		t.Error("monthly_total should be allowed even when daily_total just alerted")
	}
}
