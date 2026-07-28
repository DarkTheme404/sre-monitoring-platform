package sli

import (
	"math"
	"time"
)

// SLI — Service Level Indicator.
type SLI struct {
	name       string
	windows    []Window
	errorBudget float64 // допустимый % ошибок (0.1 = 99.9% SLA)
}

// Window — окно наблюдения для SLI.
type Window struct {
	Start    time.Time
	End      time.Time
	Total    int64
	Errors   int64
	LatencyP99 time.Duration
}

// NewSLI создаёт новый индикатор.
func NewSLI(name string, errorBudget float64) *SLI {
	return &SLI{
		name:        name,
		errorBudget: errorBudget,
		windows:     make([]Window, 0, 100),
	}
}

// Record записывает результат запроса.
func (s *SLI) Record(success bool, latency time.Duration) {
	now := time.Now()
	windowSize := 5 * time.Minute

	var last *Window
	if len(s.windows) > 0 {
		last = &s.windows[len(s.windows)-1]
		if now.Sub(last.Start) > windowSize {
			s.windows = append(s.windows, Window{
				Start: now,
				End:   now.Add(windowSize),
			})
			last = &s.windows[len(s.windows)-1]
		}
	} else {
		s.windows = append(s.windows, Window{
			Start: now,
			End:   now.Add(windowSize),
		})
		last = &s.windows[len(s.windows)-1]
	}

	last.Total++
	if !success {
		last.Errors++
	}
	if latency > last.LatencyP99 {
		last.LatencyP99 = latency
	}
}

// CurrentErrorRate возвращает текущий % ошибок.
func (s *SLI) CurrentErrorRate() float64 {
	if len(s.windows) == 0 {
		return 0
	}

	last := s.windows[len(s.windows)-1]
	if last.Total == 0 {
		return 0
	}

	return float64(last.Errors) / float64(last.Total) * 100
}

// ErrorBudgetRemaining возвращает оставшийся error budget в %.
func (s *SLI) ErrorBudgetRemaining() float64 {
	current := s.CurrentErrorRate()
	budget := s.errorBudget * 100 // в %

	remaining := budget - current
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsHealthy проверяет, здоров ли сервис.
func (s *SLI) IsHealthy() bool {
	return s.ErrorBudgetRemaining() > 0
}

// BurnRate вычисляет скорость сжигания error budget.
func (s *SLI) BurnRate() float64 {
	if len(s.windows) < 2 {
		return 0
	}

	prev := s.windows[len(s.windows)-2]
	curr := s.windows[len(s.windows)-1]

	if prev.Total == 0 {
		return 0
	}

	prevRate := float64(prev.Errors) / float64(prev.Total)
	currRate := float64(curr.Errors) / float64(curr.Total)

	if prevRate == 0 {
		if currRate > 0 {
			return math.MaxFloat64
		}
		return 0
	}

	return currRate / prevRate
}

// SLA возвращает текущий SLA (% успешных запросов).
func (s *SLI) SLA() float64 {
	var total, errors int64
	for _, w := range s.windows {
		total += w.Total
		errors += w.Errors
	}

	if total == 0 {
		return 100
	}

	return float64(total-errors) / float64(total) * 100
}
