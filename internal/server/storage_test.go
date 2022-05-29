package server

import (
	"sync"
	"testing"
)

func Test_storage_Save(t *testing.T) {
	type fields struct {
		metrics map[string]map[string]interface{}
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
			name: "save counter metric",
			fields: fields{
				metrics: map[string]map[string]interface{}{
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
			name: "save gauge metric",
			fields: fields{
				metrics: map[string]map[string]interface{}{
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
			name: "invalid metric value",
			fields: fields{
				metrics: map[string]map[string]interface{}{
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
			name: "invalid metric type",
			fields: fields{
				metrics: map[string]map[string]interface{}{
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
