package server

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

type serverConfig struct {
	address string
	port    string
}

type OptionServer func(*serverConfig) error

var ErrInvalidConfigOps = errors.New("invalid configuration option")

func newServerConfig() *serverConfig {

	const (
		address = "127.0.0.1"
		port    = "8080"
	)

	return &serverConfig{
		address: address,
		port:    port,
	}
}

func WithAddr(address string) OptionServer {
	return func(cfg *serverConfig) error {

		if r := net.ParseIP(address); r == nil {
			return fmt.Errorf("WithAddr: %w: %s", ErrInvalidConfigOps, address)
		}

		cfg.address = address
		return nil
	}
}

func WithPort(port string) OptionServer {
	return func(cfg *serverConfig) error {
		_, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("WithPort: %w: %s", ErrInvalidConfigOps, port)
		}
		cfg.port = port
		return nil
	}
}
