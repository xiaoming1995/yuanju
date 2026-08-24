package repository

import (
	"strings"
	"testing"
)

func TestDayunSummaryCacheLookupUsesCurrentChartOnly(t *testing.T) {
	for _, sql := range []string{getDayunSummarySQL, listDayunSummariesSQL} {
		for _, want := range []string{
			"s.chart_id = $1",
			"s.algorithm_version =",
			"s.years IS NOT NULL",
		} {
			if !strings.Contains(sql, want) {
				t.Fatalf("dayun summary lookup SQL should contain %q, got:\n%s", want, sql)
			}
		}
		for _, forbidden := range []string{
			"target_chart",
			"chart_hash",
			"JOIN bazi_charts",
			"ROW_NUMBER()",
		} {
			if strings.Contains(sql, forbidden) {
				t.Fatalf("new chart records must not reuse same-bazi cache; SQL should not contain %q:\n%s", forbidden, sql)
			}
		}
	}
}

func TestCurrentAlgorithmVersionBumpsForHumanizedYearNarrative(t *testing.T) {
	if CurrentAlgorithmVersion != "v3.4-ai-human-fallback" {
		t.Fatalf("CurrentAlgorithmVersion should invalidate older prompt caches, got %q", CurrentAlgorithmVersion)
	}
	if len(CurrentAlgorithmVersion) > 32 {
		t.Fatalf("CurrentAlgorithmVersion must fit ai_dayun_summaries.algorithm_version varchar(32), got length %d", len(CurrentAlgorithmVersion))
	}
}
