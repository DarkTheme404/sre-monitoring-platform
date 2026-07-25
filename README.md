# SRE Monitoring Platform

Go-based SRE monitoring and alerting platform implementing Google SRE practices.

## Architecture

```
┌─────────────┐     ┌─────────────────┐     ┌──────────────┐
│  Collector   │────▶│  SLI Tracker    │────▶│  Alerting    │
│  (scrape)   │     │  (windows)      │     │  Engine      │
└─────────────┘     └─────────────────┘     └──────┬───────┘
                                                   │
                    ┌──────────────────────────────┤
                    │                              │
               ┌────▼─────┐              ┌─────────▼────────┐
               │  HTTP    │              │  Prometheus      │
               │  API     │              │  /metrics        │
               └──────────┘              └──────────────────┘
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
