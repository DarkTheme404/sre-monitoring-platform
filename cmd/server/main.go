package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DarkTheme404/sre-monitoring-platform/internal/alerting"
	"github.com/DarkTheme404/sre-monitoring-platform/internal/api"
	"github.com/DarkTheme404/sre-monitoring-platform/internal/collector"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	engine := alerting.NewEngine(logger)

	engine.AddRule(alerting.HighErrorRateRule(1.0))
	engine.AddRule(alerting.HighLatencyRule(500))
	engine.AddRule(alerting.ErrorBudgetBurnRule(14.4))

	engine.OnAlert(func(alert alerting.Alert) {
		logger.Warn("alert notification",
			zap.String("alert_id", alert.ID),
			zap.String("name", alert.Name),
			zap.String("severity", string(alert.Severity)),
		)
	})

	coll := collector.NewCollector(logger, engine, 15*time.Second)
	coll.Start()
	defer coll.Stop()

	handler := api.NewHandler(logger, engine, coll)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":9090"
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server starting", zap.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server stopped")
}
