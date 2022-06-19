package server

import (
	"fmt"
	"github.com/sreway/yametrics/internal/metrics"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
)

func Test_storage_Save(t *testing.T) {
	type fields struct {
		metrics StorageMetrics
	}
	type args struct {
		metricType  string
		metricName  string
		metricValue string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "save counter",
			fields: fields{
				metrics: StorageMetrics{
					"counter": make(map[string]interface{}),
				},
			},
			args: args{
				metricType:  "counter",
				metricName:  "PollCount",
				metricValue: "10",
			},
			wantErr: false,
		},

		{
			name: "save gauge",
			fields: fields{
				metrics: StorageMetrics{
					"gauge": make(map[string]interface{}),
				},
			},
			args: args{
				metricType:  "gauge",
				metricName:  "RandomValue",
				metricValue: "1.1",
			},
			wantErr: false,
		},
		{
			name: "invalid value",
			fields: fields{
				metrics: StorageMetrics{
					"gauge": make(map[string]interface{}),
				},
			},
			args: args{
				metricType:  "gauge",
				metricName:  "RandomValue",
				metricValue: "invalid",
			},
			wantErr: true,
		},

		{
			name: "invalid type",
			fields: fields{
				metrics: StorageMetrics{
					"gauge": make(map[string]interface{}),
				},
			},
			args: args{
				metricType:  "invalid",
				metricName:  "RandomValue",
				metricValue: "1.1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &storage{
				metrics: tt.fields.metrics,
				mu:      sync.RWMutex{},
			}
			if err := s.Save(tt.args.metricType, tt.args.metricName, tt.args.metricValue); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_storage_GetMetricValue(t *testing.T) {
	type fields struct {
		metrics StorageMetrics
	}
	type args struct {
		metricType string
		metricName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "get counter",
			fields: fields{
				metrics: StorageMetrics{
					"counter": {
						"PollCount": metrics.Counter(10),
					},
				},
			},
			args: args{
				metricType: "counter",
				metricName: "PollCount",
			},
			want:    metrics.Counter(10),
			wantErr: assert.NoError,
		},

		{
			name: "get gauge",
			fields: fields{
				metrics: StorageMetrics{
					"gauge": {
						"RandomValue": metrics.Gauge(10),
					},
				},
			},
			args: args{
				metricType: "gauge",
				metricName: "RandomValue",
			},
			want:    metrics.Gauge(10),
			wantErr: assert.NoError,
		},

		{
			name: "invalid type",
			fields: fields{
				metrics: StorageMetrics{
					"gauge": {
						"RandomValue": metrics.Gauge(10),
					},
				},
			},
			args: args{
				metricType: "invalid",
				metricName: "RandomValue",
			},
			want:    nil,
			wantErr: assert.Error,
		},

		{
			name: "non existent gauge",
			fields: fields{
				metrics: StorageMetrics{
					"gauge": {},
				},
			},
			args: args{
				metricType: "gauge",
				metricName: "RandomValue",
			},
			want:    nil,
			wantErr: assert.Error,
		},

		{
			name: "non existent counter",
			fields: fields{
				metrics: StorageMetrics{
					"counter": {},
				},
			},
			args: args{
				metricType: "counter",
				metricName: "PollCount",
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &storage{
				metrics: tt.fields.metrics,
				mu:      sync.RWMutex{},
			}
			got, err := s.GetMetricValue(tt.args.metricType, tt.args.metricName)
			if !tt.wantErr(t, err, fmt.Sprintf("GetValue(%v, %v)", tt.args.metricType, tt.args.metricName)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetValue(%v, %v)", tt.args.metricType, tt.args.metricName)
		})
	}
}

func Test_storage_StoreMetrics(t *testing.T) {
	type fields struct {
		metrics StorageMetrics
	}
	type args struct {
		filePath string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "save metrics",
			fields: fields{
				metrics: StorageMetrics{
					"counter": {
						"PollCount": metrics.Counter(10),
					},
				},
			},

			args: args{
				filePath: "/tmp/test.json",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &storage{
				metrics: tt.fields.metrics,
				mu:      sync.RWMutex{},
			}
			tt.wantErr(t, s.StoreMetrics(tt.args.filePath), fmt.Sprintf("StoreMetrics(%v)", tt.args.filePath))
			defer os.Remove(tt.args.filePath)
		})
	}
}

func Test_storage_LoadMetrics(t *testing.T) {
	type fields struct {
		metrics StorageMetrics
	}
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    StorageMetrics
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "load metrics",
			fields: fields{
				metrics: StorageMetrics{
					"counter": {
						"PollCount": metrics.Counter(10),
					},
				},
			},

			args: args{
				filePath: "/tmp/test.json",
			},

			want: StorageMetrics{
				"counter": {
					"PollCount": metrics.Counter(10),
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &storage{
				metrics: tt.fields.metrics,
				mu:      sync.RWMutex{},
			}

			err := s.StoreMetrics(tt.args.filePath)
			assert.NoError(t, err)
			tt.wantErr(t, s.LoadMetrics(tt.args.filePath), fmt.Sprintf("LoadMetrics(%v)", tt.args.filePath))
			defer os.Remove(tt.args.filePath)
			assert.Equal(t, tt.want, s.metrics)
		})
	}
}
