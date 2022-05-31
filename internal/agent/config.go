package agent

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

type OptionAgent func(*agentConfig) error

type agentConfig struct {
	pollInterval      time.Duration
	reportInterval    time.Duration
	serverAddr        string
	serverPort        string
	serverScheme      string
	serverContentType string
	serverURL         string
}

func newAgentConfig() *agentConfig {
	const (
		pollInterval   = 2 * time.Second
		reportInterval = 10 * time.Second
		serverAddr     = "127.0.0.1"
		serverPort     = "8080"
		serverScheme   = "http"
	)

	return &agentConfig{
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		serverAddr:     serverAddr,
		serverPort:     serverPort,
		serverScheme:   serverScheme,
		serverURL:      fmt.Sprintf("%s://%s:%s", serverScheme, serverAddr, serverPort),
	}
}

var ErrInvalidConfigOps = errors.New("invalid configuration option")

func WithPollInterval(poolInterval string) OptionAgent {
	return func(cfg *agentConfig) error {
		poolIntervalDuration, err := time.ParseDuration(poolInterval)
		if err != nil {
			return fmt.Errorf("WithPollInterval: %w: %s", ErrInvalidConfigOps, poolInterval)
		}
		cfg.pollInterval = poolIntervalDuration
		return nil
	}
}

func WithReportInterval(reportInterval string) OptionAgent {
	return func(cfg *agentConfig) error {
		reportIntervalDuration, err := time.ParseDuration(reportInterval)
		if err != nil {
			return fmt.Errorf("WithReportInterval: %w: %s", ErrInvalidConfigOps, reportInterval)
		}
		cfg.reportInterval = reportIntervalDuration
		return nil
	}
}

func WithServerAddr(serverAddr string) OptionAgent {
	return func(cfg *agentConfig) error {

		if r := net.ParseIP(serverAddr); r == nil {
			return fmt.Errorf("WithServerAddr: %w: %s", ErrInvalidConfigOps, serverAddr)
		}

		cfg.serverAddr = serverAddr
		cfg.serverURL = fmt.Sprintf("%s://%s:%s", cfg.serverScheme, serverAddr, cfg.serverPort)
		return nil
	}
}

func WithServerPort(serverPort string) OptionAgent {
	return func(cfg *agentConfig) error {

		_, err := strconv.Atoi(serverPort)

		if err != nil {
			return fmt.Errorf("WithServerPort: %w: %s", ErrInvalidConfigOps, serverPort)
		}

		cfg.serverPort = serverPort
		cfg.serverURL = fmt.Sprintf("%s://%s:%s", cfg.serverScheme, cfg.serverAddr, serverPort)
		return nil
	}
}

func WithServerScheme(serverScheme string) OptionAgent {
	return func(cfg *agentConfig) error {

		validSchemes := map[string]struct{}{
			"http": {}, "https": {},
		}

		if _, ok := validSchemes[serverScheme]; !ok {
			return fmt.Errorf("WithServerPort: %w: invalid scheme %s", ErrInvalidConfigOps, serverScheme)
		}

		cfg.serverPort = serverScheme
		cfg.serverURL = fmt.Sprintf("%s://%s:%s", serverScheme, cfg.serverAddr, cfg.serverPort)
		return nil
	}
}
