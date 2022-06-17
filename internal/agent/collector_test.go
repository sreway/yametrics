package agent

import (
	"github.com/sreway/yametrics/internal/metrics"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func Test_collector_ExposeMetrics(t *testing.T) {
	type fields struct {
		metrics *metrics.RuntimeMetrics
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "expose counter metric",
			fields: fields{
				metrics: &metrics.RuntimeMetrics{
					PollCount: 10,
				},
			},
		},

		{
			name: "expose gause metric",
			fields: fields{
				metrics: &metrics.RuntimeMetrics{
					OtherSys: 884128,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &collector{
				metrics: tt.fields.metrics,
				mu:      sync.RWMutex{},
			}
			assert.NotZero(t, c.ExposeMetrics())
		})
	}
}

func Test_collector_CollectMetrics(t *testing.T) {
	type fields struct {
		metrics *metrics.RuntimeMetrics
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "collect metric",
			fields: fields{
				metrics: new(metrics.RuntimeMetrics),
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
