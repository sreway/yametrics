package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseCounter(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    Counter
		wantErr bool
	}{
		{
			name: "success purse counter",
			args: args{
				s: "20",
			},
			want:    Counter(20),
			wantErr: false,
		},

		{
			name: "invalid purse counter",
			args: args{
				s: "none",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCounter(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCounter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseCounter() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseGause(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    Gauge
		wantErr bool
	}{
		{
			name: "success purse gauge",
			args: args{
				s: "7.7",
			},
			want:    Gauge(7.7),
			wantErr: false,
		},

		{
			name: "invalid purse gauge",
			args: args{
				s: "none",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGause(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGause() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseGause() got = %v, want %v", got, tt.want)
			}
		})
	}
}

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

			got, err := m.CalcHash(tt.args.key)
			assert.NoError(t, err)
			if got != tt.want && tt.wantErr == false {
				t.Errorf("CalcHash() got = %v, want %v", got, tt.want)
				return
			}

		})
	}
}
