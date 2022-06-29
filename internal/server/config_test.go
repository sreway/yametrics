package server

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestWithAddr(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid address",
			args: args{
				address: "127.0.0.1:8080",
			},
			wantErr: false,
		},

		{
			name: "invalid address",
			args: args{
				address: "invalid",
			},
			wantErr: true,
		},

		{
			name: "invalid port",
			args: args{
				address: "127.0.0.1:invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := newServerConfig()
			assert.NoError(t, err)
			ops := WithAddr(tt.args.address)
			err = ops(cfg)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func Test_newServerConfig(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv(tt.args.envName, tt.args.envValue)
			assert.NoError(t, err)
			defer func() {
				err = os.Unsetenv(tt.args.envName)
				assert.NoError(t, err)
			}()
			assert.NoError(t, err)
			_, err = newServerConfig()
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
