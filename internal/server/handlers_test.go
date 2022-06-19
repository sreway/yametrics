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
	"strings"
	"sync"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path, body string) (*http.Response, string) {
	reader := strings.NewReader(body)
	url := fmt.Sprintf("%s%s", ts.URL, path)
	req := httptest.NewRequest(method, url, reader)
	req.RequestURI = ""
	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	err = resp.Body.Close()
	require.NoError(t, err)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			s.initRoutes(r)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp, _ := testRequest(t, ts, tt.method, tt.path, ``)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			err := resp.Body.Close()
			require.NoError(t, err)
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
			}

			r := chi.NewRouter()
			srv.initRoutes(r)
			ts := httptest.NewServer(r)
			defer ts.Close()

			path := fmt.Sprintf("/value/%s/%s", tt.fields.metricType, tt.fields.metricName)
			resp, _ := testRequest(t, ts, tt.method, path, ``)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			err := resp.Body.Close()
			require.NoError(t, err)
		})
	}
}

func Test_server_UpdateMetricJSON(t *testing.T) {
	type want struct {
		statusCode int
	}
	type args struct {
		uri    string
		method string
		body   string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "update counter",
			args: args{
				uri:    "/update/",
				method: http.MethodPost,
				body:   `{"id":"PollCounter","type":"counter","delta":5}`,
			},

			want: want{
				statusCode: 200,
			},
		},

		{
			name: "update gauge",
			args: args{
				uri:    "/update/",
				method: http.MethodPost,
				body:   `{"id":"Alloc","type":"gauge","value":767032}`,
			},

			want: want{
				statusCode: 200,
			},
		},

		{
			name: "incorrect value",
			args: args{
				uri:    "/update/",
				method: http.MethodPost,
				body:   `{"id":"PollCounter","type":"counter","incorrect":5}`,
			},
			want: want{
				statusCode: 400,
			},
		},

		{
			name: "incorrect type",
			args: args{
				uri:    "/update/",
				method: http.MethodPost,
				body:   `{"id":"Alloc","type":"incorrect","value":767032}`,
			},
			want: want{
				statusCode: 501,
			},
		},
	}

	srv := &server{
		nil,
		NewStorage(),
		nil,
	}

	r := chi.NewRouter()
	srv.initRoutes(r)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		resp, _ := testRequest(t, ts, tt.args.method, tt.args.uri, tt.args.body)
		defer resp.Body.Close()
		assert.Equal(t, tt.want.statusCode, resp.StatusCode)
	}
}

func Test_server_MetricValueJSON(t *testing.T) {
	type want struct {
		statusCode int
	}

	type fields struct {
		metricType  string
		metricName  string
		metricValue interface{}
		metricExist bool
	}

	type args struct {
		uri    string
		method string
		body   string
	}

	tests := []struct {
		name   string
		args   args
		fields fields
		want   want
	}{
		{
			name: "counter value",
			args: args{
				uri:    "/value/",
				method: http.MethodPost,
				body:   `{"id":"PollCounter","type":"counter","delta":5}`,
			},

			fields: fields{
				metricType:  "counter",
				metricName:  "PollCounter",
				metricValue: metrics.Counter(5),
				metricExist: true,
			},
			want: want{
				statusCode: 200,
			},
		},

		{
			name: "gauge value",
			args: args{
				uri:    "/value/",
				method: http.MethodPost,
				body:   `{"id":"Alloc","type":"gauge"}`,
			},
			fields: fields{
				metricType:  "gauge",
				metricName:  "Alloc",
				metricValue: metrics.Gauge(5),
				metricExist: true,
			},
			want: want{
				statusCode: 200,
			},
		},

		{
			name: "incorrect type",
			args: args{
				uri:    "/value/",
				method: http.MethodPost,
				body:   `{"id":"Alloc","type":"incorrect"}`,
			},
			fields: fields{
				metricExist: false,
			},
			want: want{
				statusCode: 501,
			},
		},

		{
			name: "not exist value",
			args: args{
				uri:    "/value/",
				method: http.MethodPost,
				body:   `{"id":"Alloc","type":"gauge"}`,
			},
			fields: fields{
				metricExist: false,
			},
			want: want{
				statusCode: 404,
			},
		},
	}

	srv := &server{
		nil,
		NewStorage(),
		nil,
	}

	r := chi.NewRouter()
	srv.initRoutes(r)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
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
		}

		r := chi.NewRouter()
		srv.initRoutes(r)
		ts := httptest.NewServer(r)
		defer ts.Close()

		resp, _ := testRequest(t, ts, tt.args.method, tt.args.uri, tt.args.body)
		defer resp.Body.Close()
		assert.Equal(t, tt.want.statusCode, resp.StatusCode)
	}
}
