package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sreway/yametrics/internal/metrics"
	"log"
	"os"
	"sync"
)

var (
	ErrNotFoundMetric = errors.New("not found metric")
	ErrStoreMetrics   = errors.New("can't store metrics")
	ErrLoadMetrics    = errors.New("can't load metrics")
)

type (
	MemoryStorage struct {
		metrics metrics.Metrics
		mu      sync.RWMutex
	}

	Storage interface {
		Save(metric metrics.Metric) error
		GetMetric(metricType, metricID string) (*metrics.Metric, error)
		GetMetrics() metrics.Metrics
		StoreMetrics(fileObj *os.File) error
		LoadMetrics(fileObj *os.File) error
		IncrementCounter(metricID string, value int64)
	}
)

func (s *MemoryStorage) UnmarshalJSON(data []byte) error {
	tmpData := new(metrics.Metrics)
	if err := json.Unmarshal(data, &tmpData); err != nil {
		return err
	}
	return nil
}

func (s *MemoryStorage) Save(metric metrics.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	storageMetrics, err := s.metrics.GetMetrics(metric.MType)

	if err != nil {
		return fmt.Errorf("Storage_Save:%w", err)
	}

	storageMetrics[metric.ID] = metric

	return nil
}

func (s *MemoryStorage) GetMetric(metricType, metricName string) (*metrics.Metric, error) {
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

func (s *MemoryStorage) GetMetrics() metrics.Metrics {
	return s.metrics
}

func (s *MemoryStorage) StoreMetrics(fileObj *os.File) error {
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

func (s *MemoryStorage) LoadMetrics(fileObj *os.File) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := json.NewDecoder(fileObj).Decode(&s.metrics); err != nil {
		return fmt.Errorf("%w: cant't decode metrics", ErrLoadMetrics)
	}
	log.Printf("success load metrics")

	return nil
}

func (s *MemoryStorage) IncrementCounter(metricID string, value int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	*s.metrics.Counter[metricID].Delta = *s.metrics.Counter[metricID].Delta + value
}

func NewMemoryStorage() Storage {
	return &MemoryStorage{
		metrics.Metrics{
			Counter: make(map[string]metrics.Metric),
			Gauge:   make(map[string]metrics.Metric),
		},
		sync.RWMutex{},
	}
}
