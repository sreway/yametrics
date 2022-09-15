// Package collector implements and describes collects metrics
package collector

import (
	"sync"

	"github.com/sreway/yametrics/internal/metrics"
)

type (
	// Collector describes the implementation of collector
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

// CollectRuntimeMetrics implements interface method for collects runtime metrics
func (c *collector) CollectRuntimeMetrics() {
	c.metrics.CollectRuntimeMetrics()
}

// CollectUtilMetrics implements interface method for collects memory and cpu metrics
func (c *collector) CollectUtilMetrics(cpuUtilization Gauge) {
	c.metrics.CollectMemmoryMetrics()
	c.metrics.SetCPUutilization(cpuUtilization)
}

// ClearPollCounter implements interface method for clear poll counter after send metrics to the server
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
