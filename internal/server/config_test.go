package server

import (
	"github.com/stretchr/testify/assert"
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
				address: "127.0.0.1",
			},
			wantErr: false,
		},

		{
			name: "invalid address",
			args: args{
				address: "256.0.0.1",
			},
			wantErr: true,
		},

		{
			name: "nil address",
			args: args{
				address: "",
			},
			wantErr: true,
		},
	}

	cfg := newServerConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := WithAddr(tt.args.address)
			err := ops(cfg)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestWithPort(t *testing.T) {
	type args struct {
		port string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid port",
			args: args{
				port: "8080",
			},
			wantErr: false,
		},

		{
			name: "invalid port",
			args: args{
				port: "invalid",
			},
			wantErr: true,
		},
	}
	cfg := newServerConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := WithPort(tt.args.port)
			err := ops(cfg)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
