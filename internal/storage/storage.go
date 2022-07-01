package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/sreway/yametrics/internal/metrics"
	"log"
	"os"
	"sync"
)

var (
	ErrNotFoundMetric     = errors.New("not found metric")
	ErrStoreMetrics       = errors.New("can't store metrics")
	ErrLoadMetrics        = errors.New("can't load metrics")
	ErrStorageUnavailable = errors.New("storage unavailable")
)

type (
	memoryStorage struct {
		metrics metrics.Metrics
		mu      sync.RWMutex
	}

	pgStorage struct {
		connection *pgx.Conn
	}

	Storage interface {
		Save(metric metrics.Metric) error
		GetMetric(metricType, metricID string) (*metrics.Metric, error)
		GetMetrics() metrics.Metrics
		IncrementCounter(metricID string, value int64)
	}

	MemoryStorage interface {
		Storage
		LoadMetrics(fileObj *os.File) error
		StoreMetrics(fileObj *os.File) error
	}

	PgStorage interface {
		Storage
		Ping(ctx context.Context) error
		Close() error
	}
)

func (s *memoryStorage) UnmarshalJSON(data []byte) error {
	tmpData := new(metrics.Metrics)
	if err := json.Unmarshal(data, &tmpData); err != nil {
		return err
	}
	return nil
}

func (s *memoryStorage) Save(metric metrics.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	storageMetrics, err := s.metrics.GetMetrics(metric.MType)

	if err != nil {
		return fmt.Errorf("Storage_Save:%w", err)
	}

	storageMetrics[metric.ID] = metric

	return nil
}

func (s *memoryStorage) GetMetric(metricType, metricName string) (*metrics.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storageMetrics, err := s.metrics.GetMetrics(metricType)

	if err != nil {
		return nil, fmt.Errorf("Storage_GetMetric:%w", err)
	}

	metric, exist := storageMetrics[metricName]

	if !exist {
		return nil, fmt.Errorf("%s: %w", metricName, ErrNotFoundMetric)
	}

	return &metric, nil

}

func (s *memoryStorage) GetMetrics() metrics.Metrics {
	return s.metrics
}

func (s *memoryStorage) StoreMetrics(fileObj *os.File) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	err := fileObj.Truncate(0)

	if err != nil {
		return fmt.Errorf("%w cat't truncate file", ErrStoreMetrics)
	}
	_, err = fileObj.Seek(0, 0)

	if err != nil {
		return fmt.Errorf("%w cat't seek file", ErrStoreMetrics)
	}

	if err := json.NewEncoder(fileObj).Encode(s.GetMetrics()); err != nil {
		return fmt.Errorf("%w: cant't encode metrics", ErrStoreMetrics)
	}
	log.Println("success save metrics to file")

	return nil
}

func (s *memoryStorage) LoadMetrics(fileObj *os.File) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := json.NewDecoder(fileObj).Decode(&s.metrics); err != nil {
		return fmt.Errorf("%w: cant't decode metrics", ErrLoadMetrics)
	}
	log.Printf("success load metrics")

	return nil
}

func (s *memoryStorage) IncrementCounter(metricID string, value int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	*s.metrics.Counter[metricID].Delta = *s.metrics.Counter[metricID].Delta + value
}

func NewMemoryStorage() MemoryStorage {
	return &memoryStorage{
		metrics.Metrics{
			Counter: make(map[string]metrics.Metric),
			Gauge:   make(map[string]metrics.Metric),
		},
		sync.RWMutex{},
	}
}

func NewPgStorage(ctx context.Context, dsn string) (PgStorage, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("NewPgStroge error: %w", err)
	}
	log.Println("NewPgStorage: success connect database")
	return &pgStorage{
		connection: conn,
	}, nil
}

func (s *pgStorage) Save(metric metrics.Metric) error {
	return nil
}

func (s *pgStorage) GetMetric(metricType, metricID string) (*metrics.Metric, error) {
	return nil, nil
}

func (s *pgStorage) GetMetrics() metrics.Metrics {
	return metrics.Metrics{}
}

func (s *pgStorage) IncrementCounter(metricID string, value int64) {}

func (s *pgStorage) Ping(ctx context.Context) error {
	if err := s.connection.Ping(ctx); err != nil {
		fmt.Println(err)
		return fmt.Errorf("pgStorage_Ping: %w", ErrStorageUnavailable)
	}
	return nil
}

func (s *pgStorage) Close() error {
	if err := s.connection.Close(context.Background()); err != nil {
		return fmt.Errorf("pgStorage_Close: %w", err)
	}
	return nil
}
