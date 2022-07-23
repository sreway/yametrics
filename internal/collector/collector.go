package collector

import (
	"sync"

	"github.com/sreway/yametrics/internal/metrics"
)

type (
	Collector interface {
		CollectRuntimeMetrics()
		CollectUtilMetrics(cpuUtilization Gauge)
		ExposeMetrics() []metrics.Metric
		ClearPollCounter()
	}
	collector struct {
		metrics *Metrics
	}
)

func (c *collector) CollectRuntimeMetrics() {
	c.metrics.CollectRuntimeMetrics()
}

func (c *collector) CollectUtilMetrics(cpuUtilization Gauge) {
	c.metrics.CollectMemmoryMetrics()
	c.metrics.SetCPUutilization(cpuUtilization)
}

func (c *collector) ClearPollCounter() {
	c.metrics.ClearPollCounter()
}

func (c *collector) ExposeMetrics() []metrics.Metric {
	return c.metrics.ExposeMetrics()
}

func NewCollector() Collector {
	return &collector{
		&Metrics{
			mu: sync.RWMutex{},
		},
	}
}
