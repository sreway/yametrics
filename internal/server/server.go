package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

type Server interface {
	Start()
	Stop()
}

type server struct {
	httpServer *http.Server
	storage    Storage
	ctx        context.Context
	stopFunc   context.CancelFunc
	wg         *sync.WaitGroup
}

type serverConfig struct {
	address string
	port    string
}

func NewServer(config *serverConfig) Server {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	return &server{
		&http.Server{
			Addr: fmt.Sprintf("%s:%s", config.address, config.port),
		},
		NewStorage(),
		ctx,
		cancel,
		wg,
	}
}

func (s *server) Start() {
	http.HandleFunc("/update/", s.UpdateMetric)
	err := s.httpServer.ListenAndServe()
	if err != nil {
		log.Printf("server start: %v", err)
		os.Exit(1)
	}
}

func (s *server) Stop() {
	s.stopFunc()
	s.wg.Wait()
}

func NewServerConfig(addr, port string) *serverConfig {
	return &serverConfig{
		address: addr,
		port:    port,
	}
}
