package metrics

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetric_CalcHash(t *testing.T) {
	type fields struct {
		ID     string
		MType  string
		MValue string
	}

	type args struct {
		key string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "counter hash",
			fields: fields{
				ID:     "testCounter",
				MType:  "counter",
				MValue: "10",
			},
			args: args{
				"SuperSecretKey",
			},
			want:    "2aea4b37f73dc2dd8e9134464c3dff3fc70c4914379fc42c4f9ec09759cb78e8",
			wantErr: false,
		},

		{
			name: "gauge hash",
			fields: fields{
				ID:     "testGauge",
				MType:  "gauge",
				MValue: "11",
			},
			args: args{
				"SuperSecretKey",
			},
			want:    "f3e50b2a3b4dbacd24fcac746ee982f5974d84cdbbf78f4adaf261d1632d34f1",
			wantErr: false,
		},

		{
			name: "invalid key",
			fields: fields{
				ID:     "testGauge",
				MType:  "gauge",
				MValue: "11",
			},
			args: args{
				"invalid",
			},
			want:    "f3e50b2a3b4dbacd24fcac746ee982f5974d84cdbbf78f4adaf261d1632d34f1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMetric(tt.fields.ID, tt.fields.MType, tt.fields.MValue)
			assert.NoError(t, err)
			got := m.CalcHash(tt.args.key)

			if tt.wantErr {
				assert.NotEqualf(t, got, tt.want, "CalcHash() got = %s, want %s", got, tt.want)
			}

			if !tt.wantErr {
				assert.Equal(t, got, tt.want, "CalcHash() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestNewMetric(t *testing.T) {
	type args struct {
		metricID    string
		metricType  string
		metricValue string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Valid metric counter",
			args: args{
				metricID:    "testCounter",
				metricType:  "counter",
				metricValue: "10",
			},
			wantErr: assert.NoError,
		},

		{
			name: "Valid metric gauge",
			args: args{
				metricID:    "testGauge",
				metricType:  "gauge",
				metricValue: "10",
			},
			wantErr: assert.NoError,
		},

		{
			name: "Incorrect metric type",
			args: args{
				metricID:    "testCounter",
				metricType:  "incorrect",
				metricValue: "10",
			},
			wantErr: assert.Error,
		},

		{
			name: "Incorrect metric value",
			args: args{
				metricID:    "testCounter",
				metricType:  "incorrect",
				metricValue: "10",
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMetric(tt.args.metricID, tt.args.metricType, tt.args.metricValue)
			if !tt.wantErr(t, err, fmt.Sprintf("NewMetric(%v, %v, %v)", tt.args.metricID, tt.args.metricType, tt.args.metricValue)) {
				return
			}
		})
	}
}

func TestMetric_GetStrValue(t *testing.T) {
	type fields struct {
		MetricID    string
		MetricValue string
		MetricType  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "counter",
			fields: fields{
				MetricID:    "testGauge",
				MetricType:  "gauge",
				MetricValue: "1",
			},
			want: "1",
		},

		{
			name: "counter",
			fields: fields{
				MetricID:    "testCounter",
				MetricType:  "counter",
				MetricValue: "1",
			},
			want: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMetric(tt.fields.MetricID, tt.fields.MetricType, tt.fields.MetricValue)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, m.GetStrValue(), "GetStrValue()")
		})
	}
}

func TestMetric_IsCounter(t *testing.T) {
	type fields struct {
		MetricID    string
		MetricValue string
		MetricType  string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "is counter",
			fields: fields{
				MetricID:    "testMetric",
				MetricType:  "counter",
				MetricValue: "1",
			},
			want: true,
		},

		{
			name: "not counter",
			fields: fields{
				MetricID:    "testMetric",
				MetricType:  "gauge",
				MetricValue: "1",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMetric(tt.fields.MetricID, tt.fields.MetricType, tt.fields.MetricValue)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, m.IsCounter(), "IsCounter()")
		})
	}
}

func TestMetric_Valid(t *testing.T) {
	type fields struct {
		ID    string
		MType string
		Delta *int64
		Value *float64
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "correct counter",
			fields: fields{
				ID:    "testCounter",
				MType: "counter",
				Delta: func(num int64) *int64 {
					return &num
				}(10),
			},
			wantErr: assert.NoError,
		},

		{
			name: "correct gauge",
			fields: fields{
				ID:    "testGauge",
				MType: "gauge",
				Value: func(num float64) *float64 {
					return &num
				}(10),
			},
			wantErr: assert.NoError,
		},

		{
			name: "incorrect counter",
			fields: fields{
				ID:    "testCounter",
				MType: "counter",
			},
			wantErr: assert.Error,
		},

		{
			name: "incorrect gauge",
			fields: fields{
				ID:    "testGauge",
				MType: "gauge",
			},
			wantErr: assert.Error,
		},

		{
			name: "incorrect metric type",
			fields: fields{
				ID:    "testGauge",
				MType: "incorrect",
				Value: func(num float64) *float64 {
					return &num
				}(10),
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				ID:    tt.fields.ID,
				MType: tt.fields.MType,
				Delta: tt.fields.Delta,
				Value: tt.fields.Value,
			}
			tt.wantErr(t, m.Valid(), "Valid()")
		})
	}
}
