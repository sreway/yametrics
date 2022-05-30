package agent

import (
	"github.com/sreway/yametrics/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
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
		pollInterval:      0,
		reportInterval:    0,
		serverAddr:        "127.0.0.1",
		serverPort:        "8080",
		serverScheme:      "http",
		serverContentType: "text/plain",
	}
}

type want struct {
	uri        string
	statusCode int
}

func Test_agent_SendToSever(t *testing.T) {
	tests := []struct {
		name string
		args []ExposeMetric
		want want
	}{
		{
			name: "send counter",
			args: []ExposeMetric{
				{
					ID:    "PollCount",
					Type:  "Counter",
					Value: metrics.Counter(5),
				},
			},
			want: want{
				uri: "/update/counter/PollCount/5",
			},
		},

		{
			name: "send gauge",
			args: []ExposeMetric{
				{
					ID:    "OtherSys",
					Type:  "Gauge",
					Value: metrics.Gauge(884128),
				},
			},
			want: want{
				uri: "/update/gauge/OtherSys/884128",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewTestHTTPClient(func(req *http.Request) *http.Response {

				assert.Equal(t, req.URL.String(), tt.want.uri)

				return &http.Response{
					StatusCode: 200,
				}
			})

			a := &agent{
				collector:  nil,
				httpClient: *client,
				Config:     NewTestAgentConfig(),
			}

			err := a.SendToSever(tt.args)
			require.NoError(t, err)
		})
	}
}
