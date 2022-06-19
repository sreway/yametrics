package server

import (
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
	"net"
	"strconv"
	"time"
)

type (
	serverConfig struct {
		Address       string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
		StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
		StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
		Restore       bool          `env:"RESTORE" envDefault:"true"`
	}
	OptionServer func(*serverConfig) error
)

var (
	ErrInvalidConfigOps = errors.New("invalid configuration option")
	ErrInvalidConfig    = errors.New("invalid configuration")
)

func newServerConfig() (*serverConfig, error) {
	cfg := serverConfig{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("newServerConfig: %v", err)
	}

	_, port, err := net.SplitHostPort(cfg.Address)

	if err != nil {
		return nil, fmt.Errorf("newServerConfig: %w invalid address %s", ErrInvalidConfig, cfg.Address)
	}

	_, err = strconv.Atoi(port)

	if err != nil {
		return nil, fmt.Errorf("newServerConfig: %w invalid port %s", ErrInvalidConfigOps, cfg.Address)
	}

	return &cfg, nil
}

func WithAddr(address string) OptionServer {
	return func(cfg *serverConfig) error {
		_, port, err := net.SplitHostPort(address)

		if err != nil {
			return fmt.Errorf("WithAddr: %w invalid address %s", ErrInvalidConfigOps, address)
		}

		_, err = strconv.Atoi(port)

		if err != nil {
			return fmt.Errorf("WithAddr: %w invalid port %s", ErrInvalidConfigOps, address)
		}

		cfg.Address = address
		return nil
	}
}
