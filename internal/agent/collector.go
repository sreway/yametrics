package agent

import (
	"github.com/sreway/yametrics/internal/metrics"
	"reflect"
	"sync"
)

type Collector interface {
	CollectMetrics()
	ExposeMetrics() []ExposeMetric
	ClearPollCounter()
}

type collector struct {
	metrics *metrics.Metrics
	mu      sync.RWMutex
}

type ExposeMetric struct {
	ID    string
	Type  string
	Value interface{}
}

func (c *collector) CollectMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics.Collect()
}

func (c *collector) ExposeMetrics() []ExposeMetric {
	c.mu.Lock()
	defer c.mu.Unlock()

	metricsElements := reflect.ValueOf(c.metrics).Elem()
	exposeMetrics := make([]ExposeMetric, 0, metricsElements.NumField())

	for i := 0; i < metricsElements.NumField(); i++ {
		exposeMetrics = append(exposeMetrics, ExposeMetric{
			ID:    metricsElements.Type().Field(i).Name,
			Value: metricsElements.Field(i).Interface(),
			Type:  metricsElements.Field(i).Type().Name(),
		})
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
		new(metrics.Metrics),
		sync.RWMutex{},
	}
}
