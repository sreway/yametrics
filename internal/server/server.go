package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sreway/yametrics/internal/metrics"
	"github.com/sreway/yametrics/internal/storage"
)

var (
	ErrInvalidMetricHash = errors.New("invalid metric hash")
	ErrInvalidStorage    = errors.New("invalid storage")
)

type (
	Server interface {
		Start()
	}

	server struct {
		httpServer *http.Server
		storage    storage.Storage
		cfg        *serverConfig
	}
)

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
		nil,
		srvCfg,
	}, nil
}

func (s *server) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	systemSignals := make(chan os.Signal, 1)
	signal.Notify(systemSignals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	exitChan := make(chan int)

	err := s.InitStorage(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	switch t := s.storage.(type) {
	case storage.MemoryStorage:
		if s.cfg.Restore {
			err = s.loadMetrics()
			if err != nil {
				log.Println(err)
			}
		}

		if s.cfg.StoreInterval != 0 {
			go s.storeMetrics(ctx)
		}

	case storage.PgStorage:
		if err = t.ValidateSchema(SourceMigrationsURL); err != nil {
			log.Fatalln(err)
		}
	default:
		log.Fatalln(ErrInvalidStorage)
	}

	go func() {
		r := chi.NewRouter()
		r.Use(middleware.Compress(s.cfg.compressLevel, s.cfg.compressTypes...))
		s.initRoutes(r)
		s.httpServer.Handler = r

		err = s.httpServer.ListenAndServe()
		if err != nil {
			log.Printf("server start: %v", err)
			systemSignals <- syscall.SIGSTOP
		}
	}()

	go func() {
		for {
			systemSignal := <-systemSignals
			switch systemSignal {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				log.Println("signal triggered.")
				if store, ok := s.storage.(storage.MemoryStorage); ok {
					if s.cfg.StoreFile != "" {
						err = store.StoreMetrics()
						if err != nil {
							log.Println(err)
						}
					}
				}
				exitChan <- 0
			default:
				log.Println("unknown signal.")
				exitChan <- 1
			}
		}
	}()

	exitCode := <-exitChan
	cancel()

	err = s.storage.Close()

	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(exitCode)
}

func (s *server) saveMetric(ctx context.Context, metric metrics.Metric, withHash bool) error {
	err := metric.Valid()
	if err != nil {
		return fmt.Errorf("Server_saveMetric: %w", err)
	}

	if withHash {
		sign := metric.CalcHash(s.cfg.Key)

		if sign != metric.Hash {
			return fmt.Errorf("Server_saveMetric error:%w",
				metrics.NewMetricError(metric.MType, metric.ID, ErrInvalidMetricHash))
		}
	}
	switch metric.IsCounter() {
	case true:
		_, err := s.storage.GetMetric(ctx, metric.MType, metric.ID)
		if err != nil {
			switch {
			case errors.Is(err, storage.ErrNotFoundMetric):
				err = s.storage.Save(ctx, metric)
				if err != nil {
					return fmt.Errorf("Server_saveMetric error:%w", err)
				}
				return nil
			default:
				return fmt.Errorf("Server_saveMetric error:%w", err)
			}
		}

		err = s.storage.IncrementCounter(ctx, metric.ID, *metric.Delta)

		if err != nil {
			return fmt.Errorf("Server_saveMetric: %w", err)
		}

	default:
		err := s.storage.Save(ctx, metric)
		if err != nil {
			return fmt.Errorf("Server_saveMetric error:%w", err)
		}
	}

	if s.cfg != nil && s.cfg.StoreInterval == 0 {
		_ = s.storage.(storage.MemoryStorage).StoreMetrics()
	}

	return nil
}

func (s *server) getMetric(ctx context.Context, metricType, metricName string, withHash bool) (metrics.Metric, error) {
	m, err := s.storage.GetMetric(ctx, metricType, metricName)
	if err != nil {
		return metrics.Metric{}, err
	}

	if withHash {
		sign := m.CalcHash(s.cfg.Key)
		m.Hash = sign
	}

	return *m, err
}

func (s *server) getMetrics(ctx context.Context) (*metrics.Metrics, error) {
	m, err := s.storage.GetMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("Server_getMetrics: %w", err)
	}

	return m, nil
}

func (s *server) getMetricsList(ctx context.Context, withHash bool) ([]metrics.Metric, error) {
	m, err := s.storage.GetMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("Server_getMetrics: %w", err)
	}

	countMetrics := len(m.Counter) + len(m.Gauge)
	metricList := make([]metrics.Metric, 0, countMetrics)

	for _, item := range m.Counter {
		if withHash {
			sign := item.CalcHash(s.cfg.Key)
			item.Hash = sign
		}
		metricList = append(metricList, item)
	}

	for _, item := range m.Gauge {
		if withHash {
			sign := item.CalcHash(s.cfg.Key)
			item.Hash = sign
		}
		metricList = append(metricList, item)
	}

	return metricList, nil
}

func (s *server) storeMetrics(ctx context.Context) {
	tick := time.NewTicker(s.cfg.StoreInterval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			err := s.storage.(storage.MemoryStorage).StoreMetrics()
			if err != nil {
				log.Println(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *server) loadMetrics() error {
	err := s.storage.(storage.MemoryStorage).LoadMetrics()
	if err != nil {
		return fmt.Errorf("Server_loadMetrics: %w", err)
	}
	return nil
}

func (s *server) InitStorage(ctx context.Context) error {
	if s.cfg.Dsn != "" {
		storageObj, err := storage.NewPgStorage(ctx, s.cfg.Dsn)
		if err == nil {
			s.storage = storageObj
			return nil
		}

		log.Printf("Server_InitStorage: %v", err)
	}

	memStorage, err := storage.NewMemoryStorage(s.cfg.StoreFile)
	if err != nil {
		return fmt.Errorf("Server_InitStorage: %w", err)
	}

	s.storage = memStorage

	return nil
}

func (s *server) pingStorage(ctx context.Context) error {
	if err := s.storage.(storage.PgStorage).Ping(ctx); err != nil {
		return fmt.Errorf("Server_pingStorage error: %w", err)
	}
	return nil
}

func (s *server) batchMetrics(ctx context.Context, m []metrics.Metric, withHash bool) error {
	if withHash {
		for _, item := range m {
			err := item.Valid()
			if err != nil {
				return fmt.Errorf("Server_batchMetrics: %w", err)
			}
			sign := item.CalcHash(s.cfg.Key)

			if sign != item.Hash {
				return fmt.Errorf("Server_batchMetric error:%w",
					metrics.NewMetricError(item.MType, item.ID, ErrInvalidMetricHash))
			}
		}
	}

	err := s.storage.BatchMetrics(ctx, m)
	if err != nil {
		return fmt.Errorf("Server_batchMetrics: %w", err)
	}

	return nil
}
