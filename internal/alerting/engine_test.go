package alerting

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewEngine(t *testing.T) {
	logger := zap.NewNop()
	engine := NewEngine(logger)
	if engine == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestAddRule(t *testing.T) {
	logger := zap.NewNop()
	engine := NewEngine(logger)

	rule := Rule{
		Name:     "test_rule",
		Severity: SeverityWarning,
		Condition: func(metrics map[string]float64) bool {
			return true
		},
	}

	engine.AddRule(rule)

	if len(engine.rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(engine.rules))
	}
}

func TestEvaluateFiresAlert(t *testing.T) {
	logger := zap.NewNop()
	engine := NewEngine(logger)

	rule := HighErrorRateRule(1.0)
	engine.AddRule(rule)

	metrics := map[string]float64{"error_rate": 2.0}
	engine.Evaluate(metrics)

	alerts := engine.ActiveAlerts()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 active alert, got %d", len(alerts))
	}

	if alerts[0].Name != "high_error_rate" {
		t.Fatalf("expected alert name high_error_rate, got %s", alerts[0].Name)
	}

	if alerts[0].Severity != SeverityCritical {
		t.Fatalf("expected critical severity, got %s", alerts[0].Severity)
	}
}

func TestEvaluateResolvesAlert(t *testing.T) {
	logger := zap.NewNop()
	engine := NewEngine(logger)

	rule := HighErrorRateRule(1.0)
	engine.AddRule(rule)

	metrics := map[string]float64{"error_rate": 2.0}
	engine.Evaluate(metrics)

	if len(engine.ActiveAlerts()) != 1 {
		t.Fatal("expected 1 active alert")
	}

	metrics["error_rate"] = 0.5
	engine.Evaluate(metrics)

	if len(engine.ActiveAlerts()) != 0 {
		t.Fatal("expected 0 active alerts after resolution")
	}
}

func TestEvaluateNoFireWhenBelowThreshold(t *testing.T) {
	logger := zap.NewNop()
	engine := NewEngine(logger)

	rule := HighErrorRateRule(5.0)
	engine.AddRule(rule)

	metrics := map[string]float64{"error_rate": 1.0}
	engine.Evaluate(metrics)

	if len(engine.ActiveAlerts()) != 0 {
		t.Fatal("expected 0 alerts when below threshold")
	}
}

func TestEvaluateDoesNotDuplicate(t *testing.T) {
	logger := zap.NewNop()
	engine := NewEngine(logger)

	rule := HighErrorRateRule(1.0)
	engine.AddRule(rule)

	metrics := map[string]float64{"error_rate": 2.0}
	engine.Evaluate(metrics)
	engine.Evaluate(metrics)

	alerts := engine.ActiveAlerts()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert (no duplicate), got %d", len(alerts))
	}
}

func TestOnAlertCallback(t *testing.T) {
	logger := zap.NewNop()
	engine := NewEngine(logger)

	var received Alert
	engine.OnAlert(func(alert Alert) {
		received = alert
	})

	rule := HighErrorRateRule(1.0)
	engine.AddRule(rule)

	metrics := map[string]float64{"error_rate": 2.0}
	engine.Evaluate(metrics)

	if received.Name != "high_error_rate" {
		t.Fatalf("expected callback for high_error_rate, got %s", received.Name)
	}
}

func TestHighLatencyRule(t *testing.T) {
	logger := zap.NewNop()
	engine := NewEngine(logger)

	rule := HighLatencyRule(500)
	engine.AddRule(rule)

	metrics := map[string]float64{"latency_p99_ms": 600}
	engine.Evaluate(metrics)

	if len(engine.ActiveAlerts()) != 1 {
		t.Fatal("expected 1 alert for high latency")
	}

	metrics["latency_p99_ms"] = 200
	engine.Evaluate(metrics)

	if len(engine.ActiveAlerts()) != 0 {
		t.Fatal("expected 0 alerts after latency drops")
	}
}

func TestErrorBudgetBurnRule(t *testing.T) {
	logger := zap.NewNop()
	engine := NewEngine(logger)

	rule := ErrorBudgetBurnRule(14.4)
	engine.AddRule(rule)

	metrics := map[string]float64{"burn_rate": 20.0}
	engine.Evaluate(metrics)

	if len(engine.ActiveAlerts()) != 1 {
		t.Fatal("expected 1 alert for budget burn")
	}

	metrics["burn_rate"] = 5.0
	engine.Evaluate(metrics)

	if len(engine.ActiveAlerts()) != 0 {
		t.Fatal("expected 0 alerts after burn rate drops")
	}
}

func TestActiveAlertsEmpty(t *testing.T) {
	logger := zap.NewNop()
	engine := NewEngine(logger)

	alerts := engine.ActiveAlerts()
	if alerts != nil {
		t.Fatalf("expected nil or empty, got %d", len(alerts))
	}
}

func TestRuleCooldown(t *testing.T) {
	rule := HighErrorRateRule(1.0)
	if rule.Cooldown != 5*time.Minute {
		t.Fatalf("expected 5m cooldown, got %v", rule.Cooldown)
	}
}
