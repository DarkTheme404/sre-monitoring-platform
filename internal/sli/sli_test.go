package sli

import (
	"testing"
	"time"
)

func TestNewSLI(t *testing.T) {
	s := NewSLI("test_sli", 0.001)
	if s == nil {
		t.Fatal("expected non-nil SLI")
	}
	if s.name != "test_sli" {
		t.Fatalf("expected name test_sli, got %s", s.name)
	}
	if s.errorBudget != 0.001 {
		t.Fatalf("expected error budget 0.001, got %f", s.errorBudget)
	}
}

func TestRecordAndCurrentErrorRate(t *testing.T) {
	s := NewSLI("test", 0.001)

	for i := 0; i < 100; i++ {
		s.Record(true, 10*time.Millisecond)
	}
	for i := 0; i < 10; i++ {
		s.Record(false, 50*time.Millisecond)
	}

	rate := s.CurrentErrorRate()
	if rate != 10.0 {
		t.Fatalf("expected 10%% error rate, got %f", rate)
	}
}

func TestCurrentErrorRateEmpty(t *testing.T) {
	s := NewSLI("empty", 0.001)
	if s.CurrentErrorRate() != 0 {
		t.Fatal("expected 0 error rate for empty SLI")
	}
}

func TestErrorBudgetRemaining(t *testing.T) {
	s := NewSLI("test", 0.001)

	for i := 0; i < 900; i++ {
		s.Record(true, 10*time.Millisecond)
	}
	for i := 0; i < 1; i++ {
		s.Record(false, 50*time.Millisecond)
	}

	budget := s.ErrorBudgetRemaining()
	if budget <= 0 {
		t.Fatalf("expected positive remaining budget, got %f", budget)
	}
}

func TestErrorBudgetExhausted(t *testing.T) {
	s := NewSLI("test", 0.001)

	for i := 0; i < 10; i++ {
		s.Record(true, 10*time.Millisecond)
	}
	for i := 0; i < 10; i++ {
		s.Record(false, 50*time.Millisecond)
	}

	budget := s.ErrorBudgetRemaining()
	if budget != 0 {
		t.Fatalf("expected 0 remaining budget, got %f", budget)
	}
}

func TestIsHealthy(t *testing.T) {
	s := NewSLI("test", 0.001)

	for i := 0; i < 1000; i++ {
		s.Record(true, 10*time.Millisecond)
	}

	if !s.IsHealthy() {
		t.Fatal("expected healthy service")
	}
}

func TestIsUnhealthy(t *testing.T) {
	s := NewSLI("test", 0.001)

	for i := 0; i < 10; i++ {
		s.Record(true, 10*time.Millisecond)
	}
	for i := 0; i < 10; i++ {
		s.Record(false, 50*time.Millisecond)
	}

	if s.IsHealthy() {
		t.Fatal("expected unhealthy service")
	}
}

func TestBurnRateNoWindows(t *testing.T) {
	s := NewSLI("test", 0.001)
	if s.BurnRate() != 0 {
		t.Fatalf("expected 0 burn rate, got %f", s.BurnRate())
	}
}

func TestBurnRateOneWindow(t *testing.T) {
	s := NewSLI("test", 0.001)
	s.Record(true, 10*time.Millisecond)
	if s.BurnRate() != 0 {
		t.Fatalf("expected 0 burn rate with one window, got %f", s.BurnRate())
	}
}

func TestSLAEmpty(t *testing.T) {
	s := NewSLI("test", 0.001)
	if s.SLA() != 100 {
		t.Fatalf("expected 100%% SLA for empty, got %f", s.SLA())
	}
}

func TestSLAWithRecords(t *testing.T) {
	s := NewSLI("test", 0.001)

	for i := 0; i < 99; i++ {
		s.Record(true, 10*time.Millisecond)
	}
	s.Record(false, 50*time.Millisecond)

	sla := s.SLA()
	if sla != 99.0 {
		t.Fatalf("expected 99%% SLA, got %f", sla)
	}
}
