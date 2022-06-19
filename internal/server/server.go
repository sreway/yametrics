package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server interface {
	Start()
}

type server struct {
	httpServer *http.Server
	storage    Storage
	cfg        *serverConfig
}

func NewServer(opts ...OptionServer) (Server, error) {
	srvCfg, err := newServerConfig()
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		err := opt(srvCfg)
		if err != nil {
			return nil, err
		}
	}

	return &server{
		&http.Server{
			Addr: srvCfg.Address,
		},
		NewStorage(),
		srvCfg,
	}, nil
}

func (s *server) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	systemSignals := make(chan os.Signal)
	signal.Notify(systemSignals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	exitChan := make(chan int)
	wg := new(sync.WaitGroup)

	if s.cfg.Restore {
		err := s.loadMetrics()
		if err != nil {
			log.Println(err)
		}
	}

	if s.cfg.StoreInterval != 0 {
		go s.storeMetrics(ctx, wg)
	}

	go func() {
		r := chi.NewRouter()
		s.initRoutes(r)
		s.httpServer.Handler = r
		err := s.httpServer.ListenAndServe()
		if err != nil {
			log.Printf("server start: %v", err)
			systemSignals <- syscall.SIGSTOP
		}
	}()

	go func() {
		for {
			systemSignal := <-systemSignals
			switch systemSignal {
			case syscall.SIGINT:
				log.Println("signal interrupt triggered.")
				_ = s.storage.StoreMetrics(s.cfg.StoreFile)
				exitChan <- 0
			case syscall.SIGTERM:
				_ = s.storage.StoreMetrics(s.cfg.StoreFile)
				log.Println("signal terminate triggered.")
				exitChan <- 0
			case syscall.SIGQUIT:
				_ = s.storage.StoreMetrics(s.cfg.StoreFile)
				log.Println("signal quit triggered.")
				exitChan <- 0
			default:
				log.Println("unknown signal.")
				exitChan <- 1
			}
		}
	}()

	exitCode := <-exitChan
	cancel()
	wg.Wait()
	os.Exit(exitCode)
}

func (s *server) saveMetric(metricType, metricName, metricValue string) error {
	err := s.storage.Save(metricType, metricName, metricValue)
	if s.cfg != nil && s.cfg.StoreInterval == 0 {
		_ = s.storage.StoreMetrics(s.cfg.StoreFile)
	}
	return err
}

func (s *server) getMetricValue(metricType, metricName string) (interface{}, error) {
	val, err := s.storage.GetMetricValue(metricType, metricName)
	return val, err
}

func (s *server) getMetrics() map[string]map[string]interface{} {
	return s.storage.GetMetrics()
}

func (s *server) storeMetrics(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	tick := time.NewTicker(s.cfg.StoreInterval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			err := s.storage.StoreMetrics(s.cfg.StoreFile)
			if err != nil {
				log.Println(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *server) loadMetrics() error {
	err := s.storage.LoadMetrics(s.cfg.StoreFile)
	if err != nil {
		return err
	}
	return nil
}
