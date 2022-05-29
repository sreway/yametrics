package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/sreway/yametrics/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	url := fmt.Sprintf("%s%s", ts.URL, path)
	req, err := http.NewRequest(method, url, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	err = resp.Body.Close()

	if err != nil {
		t.Fatal(err)
	}

	return resp, string(respBody)
}

func Test_server_UpdateMetric(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name   string
		path   string
		method string
		want   want
	}{
		{
			name:   "send counter",
			path:   "/update/counter/PollCount/100",
			method: "POST",
			want: want{
				statusCode: 200,
			},
		},

		{
			name:   "send gauge",
			path:   "/update/gauge/RandomValue/10.8",
			method: "POST",
			want: want{
				statusCode: 200,
			},
		},

		{
			name:   "invalid value",
			path:   "/update/counter/PollCount/none",
			method: "POST",
			want: want{
				statusCode: 400,
			},
		},

		{
			name:   "invalid type",
			path:   "/update/unknown/PollCount/100",
			method: "POST",
			want: want{
				statusCode: 501,
			},
		},

		{
			name:   "invalid uri",
			path:   "/update/unknown",
			method: "POST",
			want: want{
				statusCode: 404,
			},
		},
	}

	s := &server{
		nil,
		NewStorage(),
		nil,
		nil,
		nil,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			s.initRoutes(r)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp, _ := testRequest(t, ts, tt.method, tt.path)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			// need for static tests
			err := resp.Body.Close()

			if err != nil {
				t.Fatal(err)
			}
		})
	}

}

func Test_server_MetricValue(t *testing.T) {
	type want struct {
		statusCode int
	}

	type fields struct {
		metricType  string
		metricName  string
		metricValue interface{}
		metricExist bool
	}

	tests := []struct {
		name   string
		fields fields
		method string
		want   want
	}{
		{
			name: "get counter",
			fields: fields{
				metricName:  "PollCount",
				metricType:  "counter",
				metricValue: metrics.Counter(10),
				metricExist: true,
			},
			method: "GET",
			want: want{
				statusCode: 200,
			},
		},

		{
			name: "get gauge",
			fields: fields{
				metricName:  "RandomValue",
				metricType:  "gauge",
				metricValue: metrics.Gauge(10.1),
				metricExist: true,
			},
			method: "GET",
			want: want{
				statusCode: 200,
			},
		},

		{
			name: "non existent counter",
			fields: fields{
				metricName:  "PollCount",
				metricType:  "counter",
				metricValue: metrics.Counter(10),
				metricExist: false,
			},

			method: "GET",
			want: want{
				statusCode: 404,
			},
		},

		{
			name: "non existent gauge",
			fields: fields{
				metricName:  "RandomValue",
				metricType:  "gauge",
				metricValue: nil,
				metricExist: false,
			},

			method: "GET",
			want: want{
				statusCode: 404,
			},
		},

		{
			name: "invalid type",
			fields: fields{
				metricName:  "RandomValue",
				metricType:  "unknown",
				metricValue: nil,
				metricExist: false,
			},
			method: "GET",

			want: want{
				statusCode: 501,
			},
		},

		{
			name: "invalid uri",
			fields: fields{
				metricName:  "",
				metricType:  "",
				metricValue: nil,
				metricExist: false,
			},
			method: "GET",
			want: want{
				statusCode: 404,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s storage
			if tt.fields.metricExist {
				s = storage{
					metrics: map[string]map[string]interface{}{
						tt.fields.metricType: {
							tt.fields.metricName: tt.fields.metricValue,
						},
					},
					mu: sync.RWMutex{},
				}
			} else {
				s = storage{
					metrics: map[string]map[string]interface{}{},
					mu:      sync.RWMutex{},
				}
			}

			srv := &server{
				nil,
				&s,
				nil,
				nil,
				nil,
			}

			r := chi.NewRouter()
			srv.initRoutes(r)
			ts := httptest.NewServer(r)
			defer ts.Close()

			path := fmt.Sprintf("/value/%s/%s", tt.fields.metricType, tt.fields.metricName)
			resp, _ := testRequest(t, ts, tt.method, path)

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			
			// need for static tests
			err := resp.Body.Close()

			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
