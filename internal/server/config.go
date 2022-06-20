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
		Address       string        `env:"ADDRESS"`
		StoreInterval time.Duration `env:"STORE_INTERVAL"`
		StoreFile     string        `env:"STORE_FILE"`
		Restore       bool          `env:"RESTORE"`
	}
	OptionServer func(*serverConfig) error
)

var (
	AddressDefault       = "127.0.0.1:8080"
	StoreIntervalDefault = 300 * time.Second
	RestoreDefault       = true
	StoreFileDefault     = "/tmp/devops-metrics-db.json"
	ErrInvalidConfigOps  = errors.New("invalid configuration option")
	ErrInvalidConfig     = errors.New("invalid configuration")
)

func newServerConfig() (*serverConfig, error) {
	cfg := serverConfig{
		Address:       AddressDefault,
		StoreInterval: StoreIntervalDefault,
		Restore:       RestoreDefault,
		StoreFile:     StoreFileDefault,
	}

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
