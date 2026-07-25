# SRE Monitoring Platform

Платформа мониторинга и алертинга, построенная по практикам Google SRE.

## Архитектура

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Services   │────▶│  Prometheus │────▶│  Alerting   │
│  (metrics)  │     │  (scrape)   │     │  Engine     │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                    ┌──────────────────────────┼──────────┐
                    │                          │          │
               ┌────▼────┐              ┌──────▼──────┐  │
               │  SLI/   │              │  Grafana    │  │
               │  SLO    │              │  Dashboards │  │
               └─────────┘              └─────────────┘  │
                                                        │
                                                   ┌────▼────┐
                                                   │ Telegram│
                                                   │ PagerDuty│
                                                   └─────────┘
```

## Концепции SRE

### SLI (Service Level Indicator)
- **Availability** — % успешных запросов
- **Latency** — P50/P95/P99 задержка
- **Throughput** — запросов в секунду
- **Error Rate** — % ошибок

### SLO (Service Level Objective)
- Целевой SLA: 99.9% доступности
- Error Budget: 0.1% = 43.8 минут в месяц

### Error Budget
```go
// Проверка error budget
sli := sli.NewSLI("api_availability", 0.001) // 99.9%

if sli.ErrorBudgetRemaining() < 10 {
    // Менее 10% бюджета — замедляем деплои
    alertManager.Fire("error_budget_low")
}
```

## Возможности

### Мониторинг
- SLI/SLO трекинг
- Error budget burn rate
- Multi-window alerting
- Custom dashboards

### Алертинг
- Multi-window rules (5m + 1h)
- Burn rate alerting
- Severity levels (info/warning/critical)
- Auto-resolution
- Telegram/PagerDuty интеграция

### Дашборды
- Real-time метрики
- Исторические графики
- Comparison (week-over-week)
- Capacity planning

## Правила алертинга

```go
// Критический: error rate > 1% в течение 5 минут
engine.AddRule(alerting.HighErrorRateRule(1.0))

// Предупреждение: P99 latency > 500ms
engine.AddRule(alerting.HighLatencyRule(500))

// Error budget burn rate > 14.4x (бюджет сгорит за 1 час)
engine.AddRule(alerting.ErrorBudgetBurnRule(14.4))
```

## Быстрый старт

```bash
# Запуск
go run cmd/server/main.go

# Просмотр SLI
curl http://localhost:9090/api/v1/sli/api_availability
```

## Метрики

| Метрика | Описание |
|---------|----------|
| `sli_current` | Текущий SLI |
| `slo_target` | Целевой SLO |
| `error_budget_remaining` | Оставшийся error budget |
| `burn_rate` | Скорость сжигания бюджета |
| `alerts_active` | Количество активных алертов |
