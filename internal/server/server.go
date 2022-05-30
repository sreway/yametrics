package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
)

type Server interface {
	Start()
}

type server struct {
	httpServer *http.Server
	storage    Storage
}

type serverConfig struct {
	address string
	port    string
}

func NewServer(config *serverConfig) Server {
	return &server{
		&http.Server{
			Addr: fmt.Sprintf("%s:%s", config.address, config.port),
		},
		NewStorage(),
	}
}

func (s *server) Start() {
	r := chi.NewRouter()
	s.initRoutes(r)
	s.httpServer.Handler = r

	err := s.httpServer.ListenAndServe()

	if err != nil {
		log.Printf("server start: %v", err)
		os.Exit(1)
	}
}

func (s *server) saveMetric(metricType, metricName, metricValue string) error {
	err := s.storage.Save(metricType, metricName, metricValue)
	return err
}

func (s *server) getMetricValue(metricType, metricName string) (interface{}, error) {
	val, err := s.storage.GetMetricValue(metricType, metricName)
	return val, err
}

func (s *server) getMetrics() map[string]map[string]interface{} {
	return s.storage.GetMetrics()
}

func NewServerConfig(addr, port string) *serverConfig {
	return &serverConfig{
		address: addr,
		port:    port,
	}
}
