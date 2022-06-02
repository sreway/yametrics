package agent

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
	cfg := newAgentConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := WithPollInterval(tt.args.poolInterval)
			err := ops(cfg)
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
	cfg := newAgentConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := WithReportInterval(tt.args.reportInterval)
			err := ops(cfg)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestWithServerAddr(t *testing.T) {
	type args struct {
		serverAddr string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid serverAddr",
			args: args{
				serverAddr: "127.0.0.1",
			},
			wantErr: false,
		},

		{
			name: "invalid serverAddr",
			args: args{
				serverAddr: "256.0.0.1",
			},
			wantErr: true,
		},

		{
			name: "nil serverAddr",
			args: args{
				serverAddr: "",
			},
			wantErr: true,
		},
	}
	cfg := newAgentConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := WithServerAddr(tt.args.serverAddr)
			err := ops(cfg)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestWithServerPort(t *testing.T) {
	type args struct {
		serverPort string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid serverPort",
			args: args{
				serverPort: "8080",
			},
			wantErr: false,
		},

		{
			name: "invalid serverPort",
			args: args{
				serverPort: "invalid",
			},
			wantErr: true,
		},
	}
	cfg := newAgentConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := WithServerPort(tt.args.serverPort)
			err := ops(cfg)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestWithServerScheme(t *testing.T) {
	type args struct {
		serverScheme string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid serverScheme",
			args: args{
				serverScheme: "https",
			},
			wantErr: false,
		},

		{
			name: "invalid serverScheme",
			args: args{
				serverScheme: "invalid",
			},
			wantErr: true,
		},
	}
	cfg := newAgentConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := WithServerScheme(tt.args.serverScheme)
			err := ops(cfg)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
