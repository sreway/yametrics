package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sreway/yametrics/internal/metrics"
	"github.com/sreway/yametrics/internal/storage"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path, body string) *http.Response {
	reader := strings.NewReader(body)
	url := fmt.Sprintf("%s%s", ts.URL, path)
	req := httptest.NewRequest(method, url, reader)
	req.RequestURI = ""
	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	err = resp.Body.Close()
	require.NoError(t, err)

	return resp
}

func NewTestMemoryStorage(metricID, metricType, metricValue string) (storage.Storage, error) {
	metric, err := metrics.NewMetric(metricID, metricType, metricValue)
	if err != nil {
		return nil, err
	}
	testStorage, err := storage.NewMemoryStorage("")
	if err != nil {
		return nil, err
	}

	err = testStorage.Save(context.Background(), metric)

	if err != nil {
		return nil, err
	}

	return testStorage, err
}

func Test_server_UpdateMetric(t *testing.T) {
	type want struct {
		statusCode int
	}

	type args struct {
		uri    string
		method string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "send counter",
			args: args{
				uri:    "/update/counter/PollCount/100",
				method: http.MethodPost,
			},
			want: want{
				statusCode: 200,
			},
		},

		{
			name: "send gauge",
			args: args{
				uri:    "/update/gauge/RandomValue/10.8",
				method: http.MethodPost,
			},
			want: want{
				statusCode: 200,
			},
		},

		{
			name: "invalid value",
			args: args{
				uri:    "/update/counter/PollCount/none",
				method: http.MethodPost,
			},
			want: want{
				statusCode: 400,
			},
		},

		{
			name: "invalid type",
			args: args{
				uri:    "/update/unknown/PollCount/100",
				method: http.MethodPost,
			},
			want: want{
				statusCode: 501,
			},
		},

		{
			name: "invalid uri",
			args: args{
				uri:    "/update/unknown",
				method: http.MethodPost,
			},
			want: want{
				statusCode: 404,
			},
		},
	}

	cfg, err := newServerConfig()
	assert.NoError(t, err)
	store, err := storage.NewMemoryStorage(cfg.StoreFile)
	assert.NoError(t, err)
	s := &server{
		nil,
		store,
		cfg,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			s.initRoutes(r)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp := testRequest(t, ts, tt.args.method, tt.args.uri, ``)
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

	type storageData struct {
		metricID    string
		metricType  string
		metricValue string
	}

	type args struct {
		uri    string
		method string
	}

	type fields struct {
		storageData storageData
	}

	tests := []struct {
		name   string
		args   args
		fields fields
		want   want
	}{
		{
			name: "get counter",
			args: args{
				uri:    "/value/counter/PollCount",
				method: http.MethodGet,
			},

			fields: fields{
				storageData: storageData{
					metricID:    "PollCount",
					metricType:  "counter",
					metricValue: "100",
				},
			},

			want: want{
				statusCode: 200,
			},
		},
		{
			name: "get gauge",
			args: args{
				uri:    "/value/gauge/testGauge",
				method: http.MethodGet,
			},

			fields: fields{
				storageData: storageData{
					metricID:    "testGauge",
					metricType:  "gauge",
					metricValue: "100.1",
				},
			},

			want: want{
				statusCode: 200,
			},
		},

		{
			name: "non existent counter",
			args: args{
				uri:    "/value/counter/testCounter",
				method: http.MethodGet,
			},

			want: want{
				statusCode: 404,
			},
		},

		{
			name: "non existent counter",
			args: args{
				uri:    "/value/gauge/testCounter",
				method: http.MethodGet,
			},

			want: want{
				statusCode: 404,
			},
		},

		{
			name: "invalid type",
			args: args{
				uri:    "/value/unknown/testCounter",
				method: http.MethodGet,
			},

			want: want{
				statusCode: 501,
			},
		},

		{
			name: "invalid uri",
			args: args{
				uri:    "/value/unknown",
				method: http.MethodGet,
			},

			want: want{
				statusCode: 404,
			},
		},
	}
	cfg, err := newServerConfig()
	assert.NoError(t, err)
	store, err := storage.NewMemoryStorage(cfg.StoreFile)
	assert.NoError(t, err)

	s := &server{
		nil,
		store,
		cfg,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.storageData != (storageData{}) {
				testStorage, err := NewTestMemoryStorage(tt.fields.storageData.metricID,
					tt.fields.storageData.metricType, tt.fields.storageData.metricValue)
				assert.NoError(t, err)
				s.storage = testStorage
			}

			r := chi.NewRouter()
			s.initRoutes(r)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp := testRequest(t, ts, tt.args.method, tt.args.uri, ``)
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
				body:   `{"id":"testCounter","type":"counter","delta":1}`,
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
				body:   `{"id":"testGauge","type":"gauge","value":20}`,
			},

			want: want{
				statusCode: 200,
			},
		},

		{
			name: "incorrect metric data",
			args: args{
				uri:    "/update/",
				method: http.MethodPost,
				body:   ``,
			},

			want: want{
				statusCode: 400,
			},
		},

		{
			name: "incorrect method",
			args: args{
				uri:    "/update/",
				method: http.MethodGet,
				body:   `{"id":"testGauge","type":"gauge","value":20}`,
			},

			want: want{
				statusCode: 405,
			},
		},
	}
	cfg, err := newServerConfig()
	assert.NoError(t, err)
	store, err := storage.NewMemoryStorage("")
	assert.NoError(t, err)

	s := &server{
		nil,
		store,
		cfg,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			s.initRoutes(r)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp := testRequest(t, ts, tt.args.method, tt.args.uri, tt.args.body)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}

func Test_server_MetricValueJSON(t *testing.T) {
	type want struct {
		statusCode int
	}

	type storageData struct {
		metricID    string
		metricType  string
		metricValue string
	}

	type args struct {
		uri    string
		method string
		body   string
	}

	type fields struct {
		storageData storageData
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
				body:   `{"id":"PollCounter","type":"counter"}`,
			},

			fields: fields{
				storageData: storageData{
					metricID:    "PollCounter",
					metricType:  "counter",
					metricValue: "100",
				},
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
				body:   `{"id":"testGauge","type":"gauge"}`,
			},

			fields: fields{
				storageData: storageData{
					metricID:    "testGauge",
					metricType:  "gauge",
					metricValue: "100.1",
				},
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
				body:   `{"id":"testGauge","type":"incorrect"}`,
			},

			fields: fields{
				storageData: storageData{
					metricID:    "testGauge",
					metricType:  "gauge",
					metricValue: "100.1",
				},
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
				body:   `{"id":"testGauge","type":"gauge"}`,
			},
			want: want{
				statusCode: 404,
			},
		},
	}
	cfg, err := newServerConfig()
	assert.NoError(t, err)

	s := &server{
		nil,
		nil,
		cfg,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.storageData != (storageData{}) {
				testStorage, terr := NewTestMemoryStorage(tt.fields.storageData.metricID,
					tt.fields.storageData.metricType, tt.fields.storageData.metricValue)
				assert.NoError(t, terr)
				s.storage = testStorage
			} else {
				s.storage, err = storage.NewMemoryStorage(s.cfg.StoreFile)
				assert.NoError(t, err)
			}

			r := chi.NewRouter()
			s.initRoutes(r)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp := testRequest(t, ts, tt.args.method, tt.args.uri, tt.args.body)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}

func Test_server_Ping(t *testing.T) {
	type want struct {
		statusCode int
	}

	type args struct {
		uri    string
		method string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "get ping",
			args: args{
				uri:    "/ping",
				method: http.MethodGet,
			},

			want: want{
				statusCode: 200,
			},
		},

		{
			name: "incorrect method",
			args: args{
				uri:    "/ping",
				method: http.MethodPost,
			},

			want: want{
				statusCode: 405,
			},
		},
	}
	cfg, err := newServerConfig()
	assert.NoError(t, err)
	store, err := storage.NewMemoryStorage("")
	assert.NoError(t, err)

	s := &server{
		nil,
		store,
		cfg,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			s.initRoutes(r)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp := testRequest(t, ts, tt.args.method, tt.args.uri, ``)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}

func Test_server_BatchMetrics(t *testing.T) {
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
			name: "send batch metrics",
			args: args{
				uri:    "/updates/",
				method: http.MethodPost,
				body:   `[{"id":"tGauge","type":"gauge","value":20},{"id":"tCounter","type":"counter","delta":1}]`,
			},

			want: want{
				statusCode: 200,
			},
		},

		{
			name: "send incorrect batch metrics",
			args: args{
				uri:    "/updates/",
				method: http.MethodPost,
				body:   ``,
			},

			want: want{
				statusCode: 400,
			},
		},

		{
			name: "incorrect method",
			args: args{
				uri:    "/updates/",
				method: http.MethodGet,
			},

			want: want{
				statusCode: 405,
			},
		},
	}
	cfg, err := newServerConfig()
	assert.NoError(t, err)
	store, err := storage.NewMemoryStorage("")
	assert.NoError(t, err)

	s := &server{
		nil,
		store,
		cfg,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			s.initRoutes(r)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp := testRequest(t, ts, tt.args.method, tt.args.uri, tt.args.body)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}

func Test_server_Index(t *testing.T) {
	type want struct {
		statusCode int
	}

	type args struct {
		uri    string
		method string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "get index",
			args: args{
				uri:    "/",
				method: http.MethodGet,
			},
			want: want{
				statusCode: 200,
			},
		},

		{
			name: "incorrect method",
			args: args{
				uri:    "/",
				method: http.MethodPost,
			},

			want: want{
				statusCode: 405,
			},
		},
	}
	cfg, err := newServerConfig()
	assert.NoError(t, err)
	store, err := storage.NewMemoryStorage("")
	assert.NoError(t, err)

	s := &server{
		nil,
		store,
		cfg,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			s.initRoutes(r)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp := testRequest(t, ts, tt.args.method, tt.args.uri, ``)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}
