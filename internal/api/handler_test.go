package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DarkTheme404/sre-monitoring-platform/internal/alerting"
	"github.com/DarkTheme404/sre-monitoring-platform/internal/collector"
	"go.uber.org/zap"
)

func setupTestHandler(t *testing.T) *Handler {
	t.Helper()
	logger := zap.NewNop()
	engine := alerting.NewEngine(logger)
	coll := collector.NewCollector(logger, engine, 0)
	return NewHandler(logger, engine, coll)
}

func TestHandleHealth(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	h.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "healthy" {
		t.Fatalf("expected status healthy, got %v", resp["status"])
	}
}

func TestHandleAlertsEmpty(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts", nil)
	w := httptest.NewRecorder()

	h.handleAlerts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	count := resp["count"].(float64)
	if count != 0 {
		t.Fatalf("expected 0 alerts, got %v", count)
	}
}

func TestHandleSLINotFound(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sli/nonexistent", nil)
	w := httptest.NewRecorder()

	h.handleSLI(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleSLINameRequired(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sli/", nil)
	w := httptest.NewRecorder()

	h.handleSLI(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleSLIMethodNotAllowed(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sli/test", nil)
	w := httptest.NewRecorder()

	h.handleSLI(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleAlertsMethodNotAllowed(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts", nil)
	w := httptest.NewRecorder()

	h.handleAlerts(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestRegisterRoutes(t *testing.T) {
	h := setupTestHandler(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	routes := []string{"/api/v1/sli/test", "/api/v1/alerts", "/api/v1/health", "/metrics"}
	for _, route := range routes {
		req := httptest.NewRequest(http.MethodGet, route, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Fatalf("route %s not registered", route)
		}
	}
}
