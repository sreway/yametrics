package collector

import (
	"reflect"
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
		mu      sync.RWMutex
	}
)

func (c *collector) CollectRuntimeMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics.CollectRuntimeMetrics()
}

func (c *collector) CollectUtilMetrics(cpuUtilization Gauge) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics.CollectMemmoryMetrics()
	c.SetCPUutilization(cpuUtilization)
}

func (c *collector) SetCPUutilization(cpuUtilization Gauge) {
	c.metrics.CPUutilization1 = cpuUtilization
}

func (c *collector) ExposeMetrics() []metrics.Metric {
	c.mu.Lock()
	defer c.mu.Unlock()

	metricsElements := reflect.ValueOf(c.metrics).Elem()
	exposeMetrics := make([]metrics.Metric, 0, metricsElements.NumField())

	for i := 0; i < metricsElements.NumField(); i++ {
		exposeMetric := metrics.Metric{
			ID: metricsElements.Type().Field(i).Name,
		}
		switch metricsElements.Field(i).Type().Name() {
		case "Gauge":
			metricValue := metricsElements.Field(i).Interface().(Gauge).ToFloat64()
			exposeMetric.MType = "gauge"
			exposeMetric.Value = &metricValue
		case "Counter":
			metricValue := metricsElements.Field(i).Interface().(Counter).ToInt64()
			exposeMetric.MType = "counter"
			exposeMetric.Delta = &metricValue
		}

		exposeMetrics = append(exposeMetrics, exposeMetric)
	}

	return exposeMetrics
}

func (c *collector) ClearPollCounter() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics.PollCount = 0
}

func NewCollector() Collector {
	return &collector{
		new(Metrics),
		sync.RWMutex{},
	}
}
