package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/sreway/yametrics/internal/metrics"
)

func (s *memoryStorage) UnmarshalJSON(data []byte) error {
	tmpData := new(metrics.Metrics)
	err := json.Unmarshal(data, &tmpData)
	if err != nil {
		return fmt.Errorf("Storage_UnmarshalJSON")
	}
	return nil
}

func (s *memoryStorage) Save(ctx context.Context, metric metrics.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = ctx
	storageMetrics, err := s.metrics.GetMetrics(metric.MType)
	if err != nil {
		return fmt.Errorf("Storage_Save:%w", err)
	}

	storageMetrics[metric.ID] = metric

	return nil
}

func (s *memoryStorage) GetMetric(ctx context.Context, metricType, metricName string) (*metrics.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_ = ctx
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

func (s *memoryStorage) GetMetrics(ctx context.Context) (*metrics.Metrics, error) {
	_ = ctx
	return &s.metrics, nil
}

func (s *memoryStorage) StoreMetrics() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	err := s.fileObj.Truncate(0)
	if err != nil {
		return fmt.Errorf("%w cat't truncate file", ErrStoreMetrics)
	}

	_, err = s.fileObj.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("%w cat't seek file", ErrStoreMetrics)
	}

	m, err := s.GetMetrics(context.Background())
	if err != nil {
		return fmt.Errorf("memoryStorage_GetMetrics: %w", err)
	}

	if err := json.NewEncoder(s.fileObj).Encode(m); err != nil {
		return fmt.Errorf("%w: cant't encode metrics", ErrStoreMetrics)
	}

	log.Println("success save metrics to file")

	return nil
}

func (s *memoryStorage) LoadMetrics() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := json.NewDecoder(s.fileObj).Decode(&s.metrics); err != nil {
		return fmt.Errorf("%w: cant't decode metrics", ErrLoadMetrics)
	}

	log.Printf("success load metrics")

	return nil
}

func (s *memoryStorage) IncrementCounter(ctx context.Context, metricID string, value int64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_ = ctx
	*s.metrics.Counter[metricID].Delta += value

	return nil
}

func (s *memoryStorage) Close(ctx context.Context) error {
	_ = ctx
	err := s.fileObj.Close()
	if err != nil {
		return fmt.Errorf("memoryStorage_Close: %w", err)
	}

	return nil
}

func (s *memoryStorage) BatchMetrics(ctx context.Context, m []metrics.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = ctx
	counterMetrics, err := s.metrics.GetMetrics("counter")
	if err != nil {
		return fmt.Errorf("memoryStorage_BatchMetrics: %w", err)
	}

	gaugeMetrics, err := s.metrics.GetMetrics("gauge")
	if err != nil {
		return fmt.Errorf("memoryStorage_BatchMetrics: %w", err)
	}

	for _, metric := range m {
		switch metric.MType {
		case metrics.CounterStrName:
			if _, exist := counterMetrics[metric.ID]; !exist {
				counterMetrics[metric.ID] = metric
			} else {
				*counterMetrics[metric.ID].Delta += *metric.Delta
			}
		case metrics.GaugeStrName:
			gaugeMetrics[metric.ID] = metric
		default:
			return fmt.Errorf("memoryStorage_BatchMetrics: %w",
				metrics.NewMetricError(metric.MType, metric.ID, metrics.ErrInvalidMetricType))
		}
	}

	s.metrics.Counter = counterMetrics
	s.metrics.Gauge = gaugeMetrics

	return nil
}

func (s *memoryStorage) Ping(ctx context.Context) error {
	return nil
}

func NewMemoryStorage(storageFile string) (MemoryStorage, error) {
	s := &memoryStorage{
		metrics.Metrics{
			Counter: make(map[string]metrics.Metric),
			Gauge:   make(map[string]metrics.Metric),
		},
		sync.RWMutex{},
		nil,
	}

	if storageFile != "" {
		fileObj, err := OpenStorageFile(storageFile)
		if err != nil {
			return nil, fmt.Errorf("NewMemoryStorage: %w", err)
		}
		s.fileObj = fileObj
	}

	return s, nil
}

func OpenStorageFile(path string) (*os.File, error) {
	flag := os.O_RDWR | os.O_CREATE
	fileObj, err := os.OpenFile(path, flag, 0o644)
	if err != nil {
		return nil, fmt.Errorf("NewServer: can't open file %s", path)
	}
	return fileObj, nil
}
