package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"yuanju/internal/service"
)

// fakeClock 让测试固定 Now()
type fakeClock struct{ now time.Time }

func (f fakeClock) Now() time.Time { return f.now }

// stubCleaner 记录是否被调用、用什么 cutoff，返回预设值
type stubCleaner struct {
	called     bool
	gotCutoff  time.Time
	returnRows int64
	returnErr  error
}

func (s *stubCleaner) Delete(_ context.Context, cutoff time.Time) (int64, error) {
	s.called = true
	s.gotCutoff = cutoff
	return s.returnRows, s.returnErr
}

type stubRollup struct {
	called bool
	report service.RollupReport
	err    error
}

func (s *stubRollup) Rollup(_ context.Context) (service.RollupReport, error) {
	s.called = true
	return s.report, s.err
}

func makeService(t *testing.T, cfg service.CleanupConfig) (*service.CleanupService, *stubCleaner, *stubCleaner, *stubCleaner, *stubCleaner, *stubCleaner, *stubCleaner, *stubCleaner, *stubRollup) {
	t.Helper()
	aiReports := &stubCleaner{}
	polished := &stubCleaner{}
	liunian := &stubCleaner{}
	pastEvents := &stubCleaner{}
	dayunSummary := &stubCleaner{}
	compat := &stubCleaner{}
	reqLogs := &stubCleaner{}
	rollup := &stubRollup{}

	deps := service.CleanupDeps{
		AIReports:     aiReports.Delete,
		Polished:      polished.Delete,
		Liunian:       liunian.Delete,
		PastEvents:    pastEvents.Delete,
		DayunSummary:  dayunSummary.Delete,
		CompatReports: compat.Delete,
		RequestLogs:   reqLogs.Delete,
		TokenRollup:   rollup.Rollup,
	}
	svc := service.NewCleanupServiceForTest(
		deps,
		func() (service.CleanupConfig, error) { return cfg, nil },
		fakeClock{now: time.Date(2026, 5, 18, 3, 0, 0, 0, time.UTC)},
	)
	return svc, aiReports, polished, liunian, pastEvents, dayunSummary, compat, reqLogs, rollup
}

func TestRunOnce_DisabledShortCircuits(t *testing.T) {
	svc, ar, _, _, _, _, _, _, ro := makeService(t, service.CleanupConfig{Enabled: false, RetentionDays: 90, RunHour: 3})
	rep := svc.RunOnce(context.Background())
	if ar.called {
		t.Errorf("ai_reports cleaner should NOT be called when disabled")
	}
	if ro.called {
		t.Errorf("rollup should NOT be called when disabled")
	}
	if len(rep.Tables) != 0 {
		t.Errorf("disabled run should have empty Tables, got %d", len(rep.Tables))
	}
}

func TestRunOnce_PassesCutoffFromRetentionAndClock(t *testing.T) {
	svc, ar, pol, li, pe, ds, cp, rl, _ := makeService(t, service.CleanupConfig{Enabled: true, RetentionDays: 90, RunHour: 3})
	svc.RunOnce(context.Background())

	want := time.Date(2026, 5, 18, 3, 0, 0, 0, time.UTC).Add(-90 * 24 * time.Hour)
	for _, c := range []*stubCleaner{ar, pol, li, pe, ds, cp, rl} {
		if !c.called {
			t.Errorf("cleaner not called")
		}
		if !c.gotCutoff.Equal(want) {
			t.Errorf("cutoff = %v, want %v", c.gotCutoff, want)
		}
	}
}

func TestRunOnce_ErrorIsolation(t *testing.T) {
	svc, ar, pol, _, _, _, _, _, _ := makeService(t, service.CleanupConfig{Enabled: true, RetentionDays: 90, RunHour: 3})
	ar.returnErr = errors.New("simulated DB error")

	rep := svc.RunOnce(context.Background())

	if !pol.called {
		t.Errorf("polished cleaner should still run despite ai_reports failure")
	}
	if len(rep.Tables) != 7 {
		t.Fatalf("expected 7 TableResult entries, got %d", len(rep.Tables))
	}
	if rep.Tables[0].Err == nil {
		t.Errorf("expected first table (ai_reports) to have Err set")
	}
	for i := 1; i < 7; i++ {
		if rep.Tables[i].Err != nil {
			t.Errorf("Tables[%d].Err = %v, want nil", i, rep.Tables[i].Err)
		}
	}
}

func TestRunOnce_RetentionClamp(t *testing.T) {
	// 模拟 algo_config 已经 clamp 过的极小值（这里直接传 RetentionDays=1 验证传递正确）
	svc, ar, _, _, _, _, _, _, _ := makeService(t, service.CleanupConfig{Enabled: true, RetentionDays: 1, RunHour: 3})
	svc.RunOnce(context.Background())
	want := time.Date(2026, 5, 18, 3, 0, 0, 0, time.UTC).Add(-1 * 24 * time.Hour)
	if !ar.gotCutoff.Equal(want) {
		t.Errorf("cutoff = %v, want %v", ar.gotCutoff, want)
	}
}
