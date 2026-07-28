package alerting

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Severity уровень алерта.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Alert — алерт системы мониторинга.
type Alert struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Severity    Severity          `json:"severity"`
	Message     string            `json:"message"`
	Labels      map[string]string `json:"labels"`
	StartedAt   time.Time         `json:"started_at"`
	EndedAt     *time.Time        `json:"ended_at,omitempty"`
	Resolved    bool              `json:"resolved"`
}

// Rule — правило генерации алертов.
type Rule struct {
	Name       string
	Condition  func(metrics map[string]float64) bool
	Severity   Severity
	MessageTmpl string
	Cooldown   time.Duration
}

// Engine — движок алертинга.
type Engine struct {
	logger    *zap.Logger
	rules     []Rule
	alerts    map[string]*Alert
	mu        sync.RWMutex
	callbacks []func(alert Alert)
}

// NewEngine создаёт новый движок алертинга.
func NewEngine(logger *zap.Logger) *Engine {
	return &Engine{
		logger:    logger,
		rules:     make([]Rule, 0),
		alerts:    make(map[string]*Alert),
		callbacks: make([]func(alert Alert), 0),
	}
}

// AddRule добавляет правило алертинга.
func (e *Engine) AddRule(rule Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, rule)
}

// OnAlert регистрирует callback на алерт.
func (e *Engine) OnAlert(callback func(alert Alert)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.callbacks = append(e.callbacks, callback)
}

// Evaluate проверяет все правила на основе текущих метрик.
func (e *Engine) Evaluate(metrics map[string]float64) {
	e.mu.RLock()
	rules := make([]Rule, len(e.rules))
	copy(rules, e.rules)
	e.mu.RUnlock()

	for _, rule := range rules {
		if rule.Condition(metrics) {
			e.fireAlert(rule)
		} else {
			e.resolveAlert(rule.Name)
		}
	}
}

// fireAlert генерирует алерт если он ещё не активен.
func (e *Engine) fireAlert(rule Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if existing, ok := e.alerts[rule.Name]; ok && !existing.Resolved {
		return
	}

	alert := Alert{
		ID:        fmt.Sprintf("%s-%d", rule.Name, time.Now().UnixNano()),
		Name:      rule.Name,
		Severity:  rule.Severity,
		Message:   rule.MessageTmpl,
		Labels:    make(map[string]string),
		StartedAt: time.Now(),
		Resolved:  false,
	}

	e.alerts[alert.ID] = &alert

	e.logger.Warn("alert fired",
		zap.String("alert_id", alert.ID),
		zap.String("name", alert.Name),
		zap.String("severity", string(alert.Severity)),
	)

	for _, cb := range e.callbacks {
		cb(alert)
	}
}

// resolveAlert помечает алерт как resolved.
func (e *Engine) resolveAlert(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, alert := range e.alerts {
		if alert.Name == name && !alert.Resolved {
			now := time.Now()
			alert.EndedAt = &now
			alert.Resolved = true

			e.logger.Info("alert resolved",
				zap.String("alert_id", alert.ID),
				zap.String("name", alert.Name),
			)

			for _, cb := range e.callbacks {
				cb(*alert)
			}
		}
	}
}

// ActiveAlerts возвращает список активных алертов.
func (e *Engine) ActiveAlerts() []Alert {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []Alert
	for _, alert := range e.alerts {
		if !alert.Resolved {
			result = append(result, *alert)
		}
	}
	return result
}

// HighErrorRateRule — алерт при высоком % ошибок.
func HighErrorRateRule(threshold float64) Rule {
	return Rule{
		Name:       "high_error_rate",
		Severity:   SeverityCritical,
		MessageTmpl: "Error rate exceeds threshold",
		Cooldown:   5 * time.Minute,
		Condition: func(metrics map[string]float64) bool {
			errRate, ok := metrics["error_rate"]
			return ok && errRate > threshold
		},
	}
}

// HighLatencyRule — алерт при высокой задержке.
func HighLatencyRule(p99ThresholdMs float64) Rule {
	return Rule{
		Name:       "high_latency_p99",
		Severity:   SeverityWarning,
		MessageTmpl: "P99 latency exceeds threshold",
		Cooldown:   5 * time.Minute,
		Condition: func(metrics map[string]float64) bool {
			latency, ok := metrics["latency_p99_ms"]
			return ok && latency > p99ThresholdMs
		},
	}
}

// ErrorBudgetBurnRule — алерт при быстром сгорании error budget.
func ErrorBudgetBurnRule(burnRateThreshold float64) Rule {
	return Rule{
		Name:       "error_budget_burn",
		Severity:   SeverityCritical,
		MessageTmpl: "Error budget burning too fast",
		Cooldown:   10 * time.Minute,
		Condition: func(metrics map[string]float64) bool {
			burnRate, ok := metrics["burn_rate"]
			return ok && burnRate > burnRateThreshold
		},
	}
}
