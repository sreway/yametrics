package agent

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sreway/yametrics/internal/metrics"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestHTTPClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func NewTestAgentConfig() *agentConfig {
	return &agentConfig{
		PollInterval:   0,
		ReportInterval: 0,
	}
}

func IntAsPointer(value int64) *int64 {
	return &value
}

func FloatAsPointer(value float64) *float64 {
	return &value
}

func Test_agent_SendToSever(t *testing.T) {
	tests := []struct {
		name string
		args []metrics.Metric
	}{
		{
			name: "send counter",
			args: []metrics.Metric{
				{
					ID:    "PollCounter",
					MType: "counter",
					Value: FloatAsPointer(10),
				},
			},
		},

		{
			name: "send gauge",
			args: []metrics.Metric{
				{
					ID:    "PollCounter",
					MType: "counter",
					Delta: IntAsPointer(2),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewTestHTTPClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: 200,
				}
			})

			a := &agent{
				collector:  nil,
				httpClient: *client,
				Config:     NewTestAgentConfig(),
			}

			err := a.SendToSever(tt.args, false)
			require.NoError(t, err)
		})
	}
}
