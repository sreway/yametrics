package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sreway/yametrics/internal/metrics"
	"github.com/sreway/yametrics/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var ErrInvalidMetricHash = errors.New("invalid metric hash")

type Server interface {
	Start()
}

type server struct {
	httpServer  *http.Server
	storage     storage.Storage
	cfg         *serverConfig
	storageFile *os.File
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

	storageFile, err := OpenStorageFile(srvCfg.StoreFile)

	if err != nil {
		return nil, err
	}

	return &server{
		&http.Server{
			Addr: srvCfg.Address,
		},
		storage.NewMemoryStorage(),
		srvCfg,
		storageFile,
	}, nil
}

func (s *server) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	systemSignals := make(chan os.Signal)
	signal.Notify(systemSignals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	exitChan := make(chan int)

	if s.cfg.Restore {
		err := s.loadMetrics()
		if err != nil {
			log.Println(err)
		}
	}

	if s.cfg.StoreInterval != 0 {
		go s.storeMetrics(ctx)
	}

	go func() {
		r := chi.NewRouter()
		r.Use(middleware.Compress(s.cfg.compressLevel, s.cfg.compressTypes...))
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
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				log.Println("signal triggered.")
				_ = s.storage.StoreMetrics(s.storageFile)
				exitChan <- 0
			default:
				log.Println("unknown signal.")
				exitChan <- 1
			}
		}
	}()

	exitCode := <-exitChan
	cancel()
	err := s.storageFile.Close()
	if err != nil {
		panic(err)
	}
	os.Exit(exitCode)
}

func (s *server) saveMetric(metric metrics.Metric, withHash bool) error {
	if withHash {
		sign, err := metric.CalcHash(s.cfg.Key)

		if err != nil {
			return fmt.Errorf("Server_saveMetric error:%w", err)
		}

		if sign != metric.Hash {
			return fmt.Errorf("Server_saveMetric error:%w", ErrInvalidMetricHash)
		}
	}
	switch metric.IsCounter() {
	case true:
		_, err := s.storage.GetMetric(metric.MType, metric.ID)
		if err != nil {
			switch {
			case errors.Is(err, storage.ErrNotFoundMetric):
				err := s.storage.Save(metric)
				if err != nil {
					return fmt.Errorf("Server_saveMetric error:%w", err)
				}
				return nil
			default:
				return fmt.Errorf("Server_saveMetric error:%w", err)
			}
		}

		s.storage.IncrementCounter(metric.ID, *metric.Delta)

	default:
		err := s.storage.Save(metric)
		if err != nil {
			return fmt.Errorf("Server_saveMetric error:%w", err)
		}
	}

	if s.cfg != nil && s.cfg.StoreInterval == 0 {
		_ = s.storage.StoreMetrics(s.storageFile)
	}

	return nil
}

func (s *server) getMetric(metricType, metricName string, withHash bool) (metrics.Metric, error) {
	m, err := s.storage.GetMetric(metricType, metricName)
	if err != nil {
		return metrics.Metric{}, err
	}

	if withHash {
		sign, err := m.CalcHash(s.cfg.Key)

		if err != nil {
			return metrics.Metric{}, fmt.Errorf("Server_getMetric error:%w", err)
		}
		m.Hash = sign
	}

	return *m, err
}

func (s *server) getMetrics() metrics.Metrics {
	return s.storage.GetMetrics()
}

func (s *server) storeMetrics(ctx context.Context) {
	tick := time.NewTicker(s.cfg.StoreInterval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			err := s.storage.StoreMetrics(s.storageFile)
			if err != nil {
				log.Println(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *server) loadMetrics() error {
	err := s.storage.LoadMetrics(s.storageFile)
	if err != nil {
		return err
	}
	return nil
}

func OpenStorageFile(path string) (*os.File, error) {
	flag := os.O_RDWR | os.O_CREATE
	fileObj, err := os.OpenFile(path, flag, 0644)
	if err != nil {
		return nil, fmt.Errorf("NewServer: can't open file %s", path)
	}
	return fileObj, nil
}
