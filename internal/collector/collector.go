package collector

import (
	"sync"
	"time"

	"github.com/DarkTheme404/sre-monitoring-platform/internal/alerting"
	"github.com/DarkTheme404/sre-monitoring-platform/internal/sli"
	"go.uber.org/zap"
)

type MetricsSnapshot struct {
	ErrorRate   float64
	LatencyP99  float64
	Throughput  float64
	BurnRate    float64
	Healthy     bool
}

type ServiceMetrics interface {
	Scrape() (MetricsSnapshot, error)
}

type Collector struct {
	logger       *zap.Logger
	slis         map[string]*sli.SLI
	alertEngine *alerting.Engine
	services    []ServiceMetrics
	interval    time.Duration
	mu          sync.RWMutex
	snapshots   map[string]MetricsSnapshot
	stopCh      chan struct{}
}

func NewCollector(logger *zap.Logger, engine *alerting.Engine, interval time.Duration) *Collector {
	return &Collector{
		logger:      logger,
		slis:        make(map[string]*sli.SLI),
		alertEngine: engine,
		services:    make([]ServiceMetrics, 0),
		interval:    interval,
		snapshots:   make(map[string]MetricsSnapshot),
		stopCh:      make(chan struct{}),
	}
}

func (c *Collector) RegisterService(name string, svc ServiceMetrics) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services = append(c.services, svc)
	c.slis[name] = sli.NewSLI(name, 0.001)
}

func (c *Collector) GetSLI(name string) *sli.SLI {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.slis[name]
}

func (c *Collector) AllSLIs() map[string]*sli.SLI {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]*sli.SLI, len(c.slis))
	for k, v := range c.slis {
		result[k] = v
	}
	return result
}

func (c *Collector) Snapshot(name string) (MetricsSnapshot, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.snapshots[name]
	return s, ok
}

func (c *Collector) Start() {
	c.logger.Info("starting metrics collector", zap.Duration("interval", c.interval))
	go c.loop()
}

func (c *Collector) Stop() {
	close(c.stopCh)
	c.logger.Info("metrics collector stopped")
}

func (c *Collector) loop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	c.scrape()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.scrape()
		}
	}
}

func (c *Collector) scrape() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, svc := range c.services {
		snap, err := svc.Scrape()
		if err != nil {
			c.logger.Error("failed to scrape service", zap.Error(err))
			continue
		}

		metrics := map[string]float64{
			"error_rate":     snap.ErrorRate,
			"latency_p99_ms": snap.LatencyP99,
			"throughput":     snap.Throughput,
			"burn_rate":      snap.BurnRate,
		}

		c.alertEngine.Evaluate(metrics)
	}
}
