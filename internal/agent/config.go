package agent

import (
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
	"net"
	"strconv"
	"time"
)

type (
	agentConfig struct {
		PollInterval   time.Duration `env:"POLL_INTERVAL"`
		ReportInterval time.Duration `env:"REPORT_INTERVAL"`
		ServerAddress  string        `env:"ADDRESS"`
		metricEndpoint string
	}
	OptionAgent func(*agentConfig) error
)

var (
	ServerAddressDefault  = "127.0.0.1:8080"
	ReportIntervalDefault = 10 * time.Second
	PollIntervalDefault   = 2 * time.Second
	ErrInvalidConfigOps   = errors.New("invalid configuration option")
	ErrInvalidConfig      = errors.New("invalid configuration")
)

func newAgentConfig() (*agentConfig, error) {
	cfg := agentConfig{
		ServerAddress:  ServerAddressDefault,
		ReportInterval: ReportIntervalDefault,
		PollInterval:   PollIntervalDefault,
	}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("newAgentConfig: %v", err)
	}
	_, port, err := net.SplitHostPort(cfg.ServerAddress)

	if err != nil {
		return nil, fmt.Errorf("newAgentConfig: %w invalid address %s", ErrInvalidConfig, cfg.ServerAddress)
	}

	_, err = strconv.Atoi(port)

	if err != nil {
		return nil, fmt.Errorf("newAgentConfig: %w invalid port %s", ErrInvalidConfigOps, cfg.ServerAddress)
	}

	cfg.metricEndpoint = fmt.Sprintf("http://%s/update/", cfg.ServerAddress)
	return &cfg, nil
}

func WithPollInterval(poolInterval string) OptionAgent {
	return func(cfg *agentConfig) error {
		poolIntervalDuration, err := time.ParseDuration(poolInterval)
		if err != nil {
			return fmt.Errorf("WithPollInterval: %w: %s", ErrInvalidConfigOps, poolInterval)
		}
		cfg.PollInterval = poolIntervalDuration
		return nil
	}
}

func WithReportInterval(reportInterval string) OptionAgent {
	return func(cfg *agentConfig) error {
		reportIntervalDuration, err := time.ParseDuration(reportInterval)
		if err != nil {
			return fmt.Errorf("WithReportInterval: %w: %s", ErrInvalidConfigOps, reportInterval)
		}
		cfg.ReportInterval = reportIntervalDuration
		return nil
	}
}
