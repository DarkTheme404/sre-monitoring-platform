# SRE Monitoring Platform

Go-based SRE monitoring and alerting platform implementing Google SRE practices.

## Architecture

```
┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│                              SRE MONITORING PLATFORM                                          │
├──────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                              │
│   DATA SOURCES                    COLLECTION & PROCESSING               OUTPUT LAYER        │
│                                                                                              │
│   ┌──────────────┐               ┌────────────────────────┐          ┌──────────────────┐    │
│   │              │               │                        │          │                  │    │
│   │  Service A   │──────────────▶│    Metrics Collector   │          │  HTTP API        │    │
│   │  :8080       │  /metrics     │    (Periodic Scrape)   │          │  :9090           │    │
│   └──────────────┘               │                        │          │                  │    │
│                                  │  ┌──────────────────┐ │          │  GET /api/v1/    │    │
│   ┌──────────────┐               │  │  Scrape Interval │ │          │    sli/{name}    │    │
│   │              │──────────────▶│  │  (configurable)  │ │          │  GET /api/v1/    │    │
│   │  Service B   │  /metrics     │  └────────┬─────────┘ │          │    alerts        │    │
│   │  :8081       │               │           │           │          │  GET /api/v1/    │    │
│   └──────────────┘               │           │           │          │    health        │    │
│                                  │           ▼           │          │  GET /metrics    │    │
│   ┌──────────────┐               │  ┌──────────────────┐ │          └──────────────────┘    │
│   │              │──────────────▶│  │  ServiceMetrics  │ │                                   │
│   │  Service C   │  /metrics     │  │  Interface       │ │          ┌──────────────────┐    │
│   │  :8082       │               │  │  ────────────────│ │          │                  │    │
│   └──────────────┘               │  │  Success Count   │ │          │   Prometheus     │    │
│                                  │  │  Failure Count   │ │          │   :9091/metrics  │    │
│   ┌──────────────┐               │  │  Latency P50     │ │          │                  │    │
│   │              │──────────────▶│  │  Latency P95     │─┼─────────▶│  sre_collector_  │    │
│   │  External    │  /metrics     │  │  Latency P99     │ │          │  scrape_duration │    │
│   │  APIs        │               │  └──────────────────┘ │          │  sre_sli_value   │    │
│   └──────────────┘               └─────────────┬──────────┘          │  sre_alerts_     │    │
│                                                │                     │  active           │    │
│                                                │ metrics             │                  │    │
│                                                ▼                     └──────────────────┘    │
│                                  ┌──────────────────────────────────────────┐                │
│                                  │         SLI TRACKER                     │                │
│                                  │  ┌────────────────────────────────────┐  │                │
│                                  │  │                                    │  │                │
│                                  │  │  ┌──────────────────────────────┐  │  │                │
│                                  │  │  │     SLI CALCULATOR           │  │  │                │
│                                  │  │  │     (Window-Based)           │  │  │                │
│                                  │  │  │                              │  │  │                │
│                                  │  │  │  Success Rate =              │  │  │                │
│                                  │  │  │    success / total           │  │  │                │
│                                  │  │  │                              │  │  │                │
│                                  │  │  │  Latency SLI =               │  │  │                │
│                                  │  │  │    requests < threshold      │  │  │                │
│                                  │  │  │    / total requests          │  │  │                │
│                                  │  │  │                              │  │  │                │
│                                  │  │  │  Output: SLIValue           │  │  │                │
│                                  │  │  │  ┌────────────────────────┐  │  │  │                │
│                                  │  │  │  │  service: string       │  │  │  │                │
│                                  │  │  │  │  value: 0.0 - 1.0     │  │  │  │                │
│                                  │  │  │  │  window: 5m / 30m / 1h │  │  │  │                │
│                                  │  │  │  │  timestamp: time.Time  │  │  │  │                │
│                                  │  │  │  └────────────────────────┘  │  │  │                │
│                                  │  │  └──────────────────────────────┘  │  │                │
│                                  │  │              │                     │  │                │
│                                  │  │              ▼                     │  │                │
│                                  │  │  ┌──────────────────────────────┐  │  │                │
│                                  │  │  │     ERROR BUDGET            │  │  │                │
│                                  │  │  │     Calculator              │  │  │                │
│                                  │  │  │                              │  │  │                │
│                                  │  │  │  Budget = 1 - SLO           │  │  │                │
│                                  │  │  │  Consumed = errors / total  │  │  │                │
│                                  │  │  │  Remaining = budget         │  │  │                │
│                                  │  │  │             - consumed      │  │  │                │
│                                  │  │  │                              │  │  │                │
│                                  │  │  │  Burn Rate =                │  │  │                │
│                                  │  │  │  consumed / budget          │  │  │                │
│                                  │  │  │  (across windows)           │  │  │                │
│                                  │  │  └──────────────────────────────┘  │  │                │
│                                  │  │                                    │  │                │
│                                  │  └────────────────────────────────────┘  │                │
│                                  └─────────────────────┬────────────────────┘                │
│                                                        │                                     │
│                                                        │ SLI values                          │
│                                                        ▼                                     │
│                                  ┌──────────────────────────────────────────┐                │
│                                  │         ALERT ENGINE                     │                │
│                                  │  ┌────────────────────────────────────┐  │                │
│                                  │  │                                    │  │                │
│                                  │  │  ┌──────────────────────────────┐  │  │                │
│                                  │  │  │     RULE EVALUATOR           │  │  │                │
│                                  │  │  │                              │  │  │                │
│                                  │  │  │  Rules:                      │  │  │                │
│                                  │  │  │  ┌────────────────────────┐  │  │  │                │
│                                  │  │  │  │ HighErrorRateRule     │  │  │  │                │
│                                  │  │  │  │ error_rate > 5%       │  │  │  │                │
│                                  │  │  │  │ severity: CRITICAL    │  │  │  │                │
│                                  │  │  │  └────────────────────────┘  │  │  │                │
│                                  │  │  │  ┌────────────────────────┐  │  │  │                │
│                                  │  │  │  │ HighLatencyRule       │  │  │  │                │
│                                  │  │  │  │ p99 > 500ms           │  │  │  │                │
│                                  │  │  │  │ severity: WARNING     │  │  │  │                │
│                                  │  │  │  └────────────────────────┘  │  │  │                │
│                                  │  │  │  ┌────────────────────────┐  │  │  │                │
│                                  │  │  │  │ ErrorBudgetBurnRule   │  │  │  │                │
│                                  │  │  │  │ burn_rate > 2.0       │  │  │  │                │
│                                  │  │  │  │ severity: CRITICAL    │  │  │  │                │
│                                  │  │  │  └────────────────────────┘  │  │  │                │
│                                  │  │  │                              │  │  │                │
│                                  │  │  └──────────────────────────────┘  │  │                │
│                                  │  │              │                     │  │                │
│                                  │  │              ▼                     │  │                │
│                                  │  │  ┌──────────────────────────────┐  │  │                │
│                                  │  │  │     SEVERITY MANAGER        │  │  │                │
│                                  │  │  │                              │  │  │                │
│                                  │  │  │  Levels:                     │  │  │                │
│                                  │  │  │  ┌────────────────────────┐  │  │  │                │
│                                  │  │  │  │  INFO     ───────▶    │  │  │  │                │
│                                  │  │  │  │  WARNING  ───────▶    │  │  │  │                │
│                                  │  │  │  │  CRITICAL ───────▶    │  │  │  │                │
│                                  │  │  │  └────────────────────────┘  │  │  │                │
│                                  │  │  │                              │  │  │                │
│                                  │  │  │  Auto-Resolution when        │  │  │                │
│                                  │  │  │  conditions clear            │  │  │                │
│                                  │  │  └──────────────────────────────┘  │  │                │
│                                  │  │                                    │  │                │
│                                  │  └────────────────────────────────────┘  │                │
│                                  └─────────────────────┬────────────────────┘                │
│                                                        │                                     │
│                                                        │ alerts                              │
│                                                        ▼                                     │
│                                  ┌──────────────────────────────────────────┐                │
│                                  │         NOTIFICATIONS                    │                │
│                                  │  ┌────────────────────────────────────┐  │                │
│                                  │  │                                    │  │                │
│                                  │  │  ┌──────────────────────────────┐  │  │                │
│                                  │  │  │  Callback System             │  │  │                │
│                                  │  │  │  ─────────────────           │  │  │                │
│                                  │  │  │  Custom webhook handlers     │  │  │                │
│                                  │  │  │  Pluggable notification      │  │  │                │
│                                  │  │  │  backends                    │  │  │                │
│                                  │  │  └──────────────────────────────┘  │  │                │
│                                  │  │                                    │  │                │
│                                  │  │  ┌──────────────────────────────┐  │  │                │
│                                  │  │  │  Integrations:               │  │  │                │
│                                  │  │  │  - Telegram Bot              │  │  │                │
│                                  │  │  │  - Slack Webhook             │  │  │                │
│                                  │  │  │  - Email (SMTP)              │  │  │                │
│                                  │  │  │  - PagerDuty                 │  │  │                │
│                                  │  │  └──────────────────────────────┘  │  │                │
│                                  │  │                                    │  │                │
│                                  │  └────────────────────────────────────┘  │                │
│                                  └──────────────────────────────────────────┘                │
│                                                                                              │
│   ┌─────────────────────────────────────────────────────────────────────────────────────────┐ │
│   │                              DATA FLOW                                                  │ │
│   │                                                                                         │ │
│   │   Services ──▶ Collector ──▶ SLI Tracker ──▶ Error Budget ──▶ Alert Engine ──▶ Notify  │ │
│   │      │                                                              │                   │ │
│   │      │                                                              ▼                   │ │
│   │      │    ◀── Health Check ◀─────────────────────────────── Prometheus                  │ │
│   │      │    ◀── SLI Query ◀── HTTP API ◀── Alert State                                    │ │
│   │      │                                                                                │ │
│   │      └── /metrics ──▶ Prometheus ──▶ Grafana Dashboards                                 │ │
│   └─────────────────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                              │
└──────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Features

### SLI/SLO Tracking
- Window-based SLI recording (success/failure + latency)
- Error budget calculation and remaining budget
- Burn rate computation across observation windows
- SLA percentage aggregation

### Alerting Engine
- Rule-based alerting with configurable conditions
- Severity levels: info, warning, critical
- Auto-resolution when conditions clear
- Callback system for notifications
- Built-in rules: `HighErrorRateRule`, `HighLatencyRule`, `ErrorBudgetBurnRule`

### Metrics Collector
- Periodic scraping of service metrics
- Feeds metrics to SLI tracker and alerting engine
- Pluggable `ServiceMetrics` interface

### HTTP API
- `GET /api/v1/sli/{name}` — SLI metrics for a named service
- `GET /api/v1/alerts` — active alerts
- `GET /api/v1/health` — health check
- `GET /metrics` — Prometheus metrics endpoint

## Quick Start

```bash
# Run
go run cmd/server/main.go

# Or with Docker
docker build -t sre-monitoring-platform .
docker run -p 9090:9090 sre-monitoring-platform

# Query SLI
curl http://localhost:9090/api/v1/sli/api_availability

# Query alerts
curl http://localhost:9090/api/v1/alerts

# Health check
curl http://localhost:9090/api/v1/health
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/sli/{name}` | GET | SLI metrics (error rate, budget, burn rate, SLA) |
| `/api/v1/alerts` | GET | List active alerts |
| `/api/v1/health` | GET | Health status |
| `/metrics` | GET | Prometheus metrics |

## Project Structure

```
cmd/server/main.go          — HTTP server with graceful shutdown
internal/api/handler.go     — HTTP handlers
internal/collector/collector.go — Metrics collector
internal/sli/sli.go         — SLI/SLO calculations
internal/alerting/engine.go — Alerting engine
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `LISTEN_ADDR` | `:9090` | Server listen address |
