package api

import (
	"encoding/json"
	"net/http"

	"github.com/DarkTheme404/sre-monitoring-platform/internal/alerting"
	"github.com/DarkTheme404/sre-monitoring-platform/internal/collector"
	"github.com/DarkTheme404/sre-monitoring-platform/internal/sli"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Handler struct {
	logger       *zap.Logger
	alertEngine  *alerting.Engine
	collector    *collector.Collector
}

func NewHandler(logger *zap.Logger, engine *alerting.Engine, coll *collector.Collector) *Handler {
	return &Handler{
		logger:      logger,
		alertEngine: engine,
		collector:   coll,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/sli/", h.handleSLI)
	mux.HandleFunc("/api/v1/alerts", h.handleAlerts)
	mux.HandleFunc("/api/v1/health", h.handleHealth)
	mux.Handle("/metrics", promhttp.Handler())
}

func (h *Handler) handleSLI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Path[len("/api/v1/sli/"):]
	if name == "" {
		http.Error(w, "sli name required", http.StatusBadRequest)
		return
	}

	slis := h.collector.AllSLIs()
	s, ok := slis[name]
	if !ok {
		http.Error(w, "sli not found", http.StatusNotFound)
		return
	}

	resp := map[string]interface{}{
		"name":                  name,
		"current_error_rate":    s.CurrentErrorRate(),
		"error_budget_remaining": s.ErrorBudgetRemaining(),
		"burn_rate":             s.BurnRate(),
		"sla":                   s.SLA(),
		"healthy":               s.IsHealthy(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	alerts := h.alertEngine.ActiveAlerts()
	if alerts == nil {
		alerts = make([]alerting.Alert, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slis := h.collector.AllSLIs()
	healthy := true
	for _, s := range slis {
		if !s.IsHealthy() {
			healthy = false
			break
		}
	}

	status := "healthy"
	if !healthy {
		status = "degraded"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  status,
		"healthy": healthy,
	})
}
