package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
