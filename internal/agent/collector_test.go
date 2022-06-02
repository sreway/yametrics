package agent

import (
	"github.com/sreway/yametrics/internal/metrics"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func Test_collector_ExposeMetrics(t *testing.T) {
	type fields struct {
		metrics *metrics.Metrics
	}
	tests := []struct {
		name   string
		fields fields
		want   ExposeMetric
	}{
		{
			name: "expose counter metric",
			fields: fields{
				metrics: &metrics.Metrics{
					PollCount: 10,
				},
			},
			want: ExposeMetric{
				ID:    "PollCount",
				Type:  "Counter",
				Value: metrics.Counter(10),
			},
		},

		{
			name: "expose gause metric",
			fields: fields{
				metrics: &metrics.Metrics{
					OtherSys: 884128,
				},
			},
			want: ExposeMetric{
				ID:    "OtherSys",
				Type:  "Gauge",
				Value: metrics.Gauge(884128),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &collector{
				metrics: tt.fields.metrics,
				mu:      sync.RWMutex{},
			}
			assert.Containsf(t, c.ExposeMetrics(), tt.want, "ExposeMetrics()")
		})
	}
}

func Test_collector_CollectMetrics(t *testing.T) {
	type fields struct {
		metrics *metrics.Metrics
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "collect metric",
			fields: fields{
				metrics: new(metrics.Metrics),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &collector{
				metrics: tt.fields.metrics,
				mu:      sync.RWMutex{},
			}
			c.CollectMetrics()
			assert.NotZero(t, c.metrics.PollCount)
			assert.NotZero(t, c.metrics.RandomValue)
			assert.NotZero(t, c.metrics.OtherSys)
		})
	}
}
