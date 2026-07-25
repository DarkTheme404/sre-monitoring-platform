package collector

import (
	"testing"
	"time"

	"github.com/DarkTheme404/sre-monitoring-platform/internal/alerting"
	"go.uber.org/zap"
)

type mockService struct {
	snapshot MetricsSnapshot
	err      error
}

func (m *mockService) Scrape() (MetricsSnapshot, error) {
	return m.snapshot, m.err
}

func TestNewCollector(t *testing.T) {
	logger := zap.NewNop()
	engine := alerting.NewEngine(logger)
	coll := NewCollector(logger, engine, 10*time.Second)

	if coll == nil {
		t.Fatal("expected non-nil collector")
	}
	if coll.interval != 10*time.Second {
		t.Fatalf("expected 10s interval, got %v", coll.interval)
	}
}

func TestRegisterService(t *testing.T) {
	logger := zap.NewNop()
	engine := alerting.NewEngine(logger)
	coll := NewCollector(logger, engine, 10*time.Second)

	svc := &mockService{
		snapshot: MetricsSnapshot{ErrorRate: 0.5, LatencyP99: 100},
	}

	coll.RegisterService("test-service", svc)

	slis := coll.AllSLIs()
	if _, ok := slis["test-service"]; !ok {
		t.Fatal("expected test-service SLI to be registered")
	}
}

func TestGetSLI(t *testing.T) {
	logger := zap.NewNop()
	engine := alerting.NewEngine(logger)
	coll := NewCollector(logger, engine, 10*time.Second)

	svc := &mockService{}
	coll.RegisterService("my-svc", svc)

	s := coll.GetSLI("my-svc")
	if s == nil {
		t.Fatal("expected non-nil SLI")
	}

	if coll.GetSLI("nonexistent") != nil {
		t.Fatal("expected nil for nonexistent SLI")
	}
}

func TestSnapshot(t *testing.T) {
	logger := zap.NewNop()
	engine := alerting.NewEngine(logger)
	coll := NewCollector(logger, engine, 10*time.Second)

	_, ok := coll.Snapshot("missing")
	if ok {
		t.Fatal("expected no snapshot for unscraped service")
	}
}

func TestScrape(t *testing.T) {
	logger := zap.NewNop()
	engine := alerting.NewEngine(logger)
	coll := NewCollector(logger, engine, 10*time.Second)

	svc := &mockService{
		snapshot: MetricsSnapshot{ErrorRate: 2.0, LatencyP99: 300, Throughput: 500},
	}
	coll.RegisterService("svc1", svc)

	coll.scrape()

	slis := coll.AllSLIs()
	s := slis["svc1"]
	if s == nil {
		t.Fatal("expected SLI for svc1")
	}

	_ = s
}

func TestScrapeError(t *testing.T) {
	logger := zap.NewNop()
	engine := alerting.NewEngine(logger)
	coll := NewCollector(logger, engine, 10*time.Second)

	svc := &mockService{
		err: assertError("scrape failed"),
	}
	coll.RegisterService("failing-svc", svc)

	coll.scrape()
}

type assertError string

func (e assertError) Error() string { return string(e) }
