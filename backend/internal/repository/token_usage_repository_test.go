package repository

import (
	"testing"
	"time"
	"yuanju/pkg/database"
)

// fakeCostFn returns 0.01 CNY per 1000 tokens (any model)
var fakeCostFn = func(_ string, p, c int) float64 {
	return float64(p+c) * 0.01 / 1000
}

// requireDB：无已初始化的 database.DB 时跳过（此前用 testing.Short() 守卫，
// 普通 go test ./... 会因 DB 为 nil 直接 panic）
func requireDB(t *testing.T) {
	t.Helper()
	if testing.Short() || database.DB == nil {
		t.Skip("requires DB (database.DB not initialized)")
	}
}

func TestGetTokenUsageCostByModel_AggregatesGroupedRows(t *testing.T) {
	requireDB(t)
	now := time.Now()
	from := now.AddDate(0, 0, -1)
	to := now

	rows, totalTokens, err := GetTokenUsageCostByModel(from, to, fakeCostFn)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if rows == nil {
		t.Error("expected non-nil rows")
	}
	if totalTokens < 0 {
		t.Errorf("totalTokens must be non-negative, got %d", totalTokens)
	}
}

func TestGetUserCostBreakdown_RespectsLimitAndSortsByCost(t *testing.T) {
	requireDB(t)
	now := time.Now()
	from := now.AddDate(0, 0, -7)
	to := now

	rows, err := GetUserCostBreakdown(from, to, 5, fakeCostFn)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(rows) > 5 {
		t.Errorf("len > 5: %d", len(rows))
	}
	for i := 1; i < len(rows); i++ {
		if rows[i-1].TotalCostCny < rows[i].TotalCostCny {
			t.Errorf("rows[%d].TotalCostCny=%f < rows[%d].TotalCostCny=%f (must be desc)",
				i-1, rows[i-1].TotalCostCny, i, rows[i].TotalCostCny)
		}
	}
}
