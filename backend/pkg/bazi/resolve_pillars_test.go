package bazi

import "testing"

func TestResolvePillars_RoundTripContainsOriginal(t *testing.T) {
	y, m, d, h := 1984, 2, 15, 12
	r := Calculate(y, m, d, h, "male", false, 0, "solar", false)
	yearGZ := r.YearGan + r.YearZhi
	monthGZ := r.MonthGan + r.MonthZhi
	dayGZ := r.DayGan + r.DayZhi
	hourGZ := r.HourGan + r.HourZhi

	cands := ResolvePillars(yearGZ, monthGZ, dayGZ, hourGZ, 1900, 2030, 2026)
	if len(cands) == 0 {
		t.Fatalf("expected at least one candidate, got 0")
	}

	found := false
	for _, c := range cands {
		if c.Year == y && c.Month == m && c.Day == d {
			found = true
		}
		cr := Calculate(c.Year, c.Month, c.Day, c.Hour, "male", false, 0, "solar", false)
		if cr.YearGan+cr.YearZhi != yearGZ ||
			cr.MonthGan+cr.MonthZhi != monthGZ ||
			cr.DayGan+cr.DayZhi != dayGZ ||
			cr.HourGan+cr.HourZhi != hourGZ {
			t.Errorf("candidate %v does not reproduce target pillars", c)
		}
	}
	if !found {
		t.Errorf("original date %d-%d-%d not found in candidates %v", y, m, d, cands)
	}
}

func TestResolvePillars_MidpointHour(t *testing.T) {
	r := Calculate(1984, 2, 15, 12, "male", false, 0, "solar", false)
	cands := ResolvePillars(r.YearGan+r.YearZhi, r.MonthGan+r.MonthZhi,
		r.DayGan+r.DayZhi, r.HourGan+r.HourZhi, 1980, 1990, 2026)
	if len(cands) == 0 {
		t.Fatalf("expected candidates")
	}
	for _, c := range cands {
		if c.Hour != 12 {
			t.Errorf("expected midpoint hour 12 for 午时, got %d", c.Hour)
		}
	}
}

func TestResolvePillars_InvalidPillarsReturnEmpty(t *testing.T) {
	cands := ResolvePillars("甲丑", "丙寅", "戊辰", "庚午", 1900, 2030, 2026)
	if len(cands) != 0 {
		t.Errorf("expected empty for invalid ganzhi, got %v", cands)
	}
}

func TestResolvePillars_InconsistentPillarsReturnEmpty(t *testing.T) {
	cands := ResolvePillars("甲子", "甲子", "甲子", "甲子", 1900, 2030, 2026)
	if len(cands) != 0 {
		t.Errorf("expected empty for non-self-consistent pillars, got %v", cands)
	}
}

func TestResolvePillars_RefAge(t *testing.T) {
	r := Calculate(1984, 2, 15, 12, "male", false, 0, "solar", false)
	cands := ResolvePillars(r.YearGan+r.YearZhi, r.MonthGan+r.MonthZhi,
		r.DayGan+r.DayZhi, r.HourGan+r.HourZhi, 1980, 1990, 2026)
	for _, c := range cands {
		if c.RefAge != 2026-c.Year {
			t.Errorf("RefAge mismatch: got %d, want %d", c.RefAge, 2026-c.Year)
		}
	}
}
