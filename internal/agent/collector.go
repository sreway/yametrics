package agent

import (
	"github.com/sreway/yametrics/internal/metrics"
	"reflect"
	"sync"
)

type Collector interface {
	CollectMetrics()
	ExposeMetrics() []metrics.Metric
	ClearPollCounter()
}

type collector struct {
	metrics *metrics.RuntimeMetrics
	mu      sync.RWMutex
}

func (c *collector) CollectMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics.Collect()
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
			metricValue := metricsElements.Field(i).Interface().(metrics.Gauge).ToFloat64()
			exposeMetric.MType = "gauge"
			exposeMetric.Value = &metricValue
		case "Counter":
			metricValue := metricsElements.Field(i).Interface().(metrics.Counter).ToInt64()
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
		new(metrics.RuntimeMetrics),
		sync.RWMutex{},
	}
}
