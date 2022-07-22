package agent

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithPollInterval(t *testing.T) {
	type args struct {
		poolInterval string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid pollInterval",
			args: args{
				poolInterval: "5s",
			},
			wantErr: false,
		},

		{
			name: "invalid pollInterval",
			args: args{
				poolInterval: "5s5",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := newAgentConfig()
			assert.NoError(t, err)
			ops := WithPollInterval(tt.args.poolInterval)
			err = ops(cfg)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestWithReportInterval(t *testing.T) {
	type args struct {
		reportInterval string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid reportInterval",
			args: args{
				reportInterval: "5s",
			},
			wantErr: false,
		},

		{
			name: "invalid reportInterval",
			args: args{
				reportInterval: "5s5",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := newAgentConfig()
			assert.NoError(t, err)
			ops := WithReportInterval(tt.args.reportInterval)
			err = ops(cfg)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func Test_newAgentConfig(t *testing.T) {
	type args struct {
		envName  string
		envValue string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid address",
			args: args{
				envName:  "ADDRESS",
				envValue: "127.0.0.1:8080",
			},
			wantErr: false,
		},

		{
			name: "invalid address",
			args: args{
				envName:  "ADDRESS",
				envValue: "invalid",
			},
			wantErr: true,
		},

		{
			name: "invalid port",
			args: args{
				envName:  "ADDRESS",
				envValue: "127.0.0.1:invalid",
			},
			wantErr: true,
		},

		{
			name: "valid poll interval",
			args: args{
				envName:  "POLL_INTERVAL",
				envValue: "2s",
			},
			wantErr: false,
		},

		{
			name: "invalid poll interval",
			args: args{
				envName:  "POLL_INTERVAL",
				envValue: "invalid",
			},
			wantErr: true,
		},

		{
			name: "valid report interval",
			args: args{
				envName:  "REPORT_INTERVAL",
				envValue: "10s",
			},
			wantErr: false,
		},

		{
			name: "invalid report interval",
			args: args{
				envName:  "REPORT_INTERVAL",
				envValue: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv(tt.args.envName, tt.args.envValue)
			defer func() {
				err = os.Unsetenv(tt.args.envName)
				assert.NoError(t, err)
			}()
			assert.NoError(t, err)
			_, err = newAgentConfig()
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
