package server

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
	ErrInvalidMetricValue = errors.New("invalid metric value")
	ErrInvalidMetricType  = errors.New("invalid metric type")
	ErrNotFoundMetric     = errors.New("not found metric")
	ErrStoreMetrics       = errors.New("can't store metrics")
	ErrLoadMetrics        = errors.New("can't load metrics")
)

type (
	StorageMetrics map[string]map[string]interface{}
	Storage        interface {
		Save(metricType, metricName, metricValue string) error
		GetMetricValue(metricType, metricName string) (interface{}, error)
		GetMetrics() StorageMetrics
		StoreMetrics(filePath string) error
		LoadMetrics(filePath string) error
	}

	storage struct {
		metrics StorageMetrics
		mu      sync.RWMutex
	}
)

func (s *storage) UnmarshalJSON(data []byte) error {
	tmpData := make(StorageMetrics)

	if err := json.Unmarshal(data, &tmpData); err != nil {
		return err
	}

	for mType, mData := range tmpData {
		for k, v := range mData {
			switch mType {
			case "gauge":
				s.metrics[mType][k] = metrics.Gauge(v.(float64))
			case "counter":
				s.metrics[mType][k] = metrics.Counter(v.(float64))

			default:
				return fmt.Errorf("incorrect input type")
			}
		}
	}

	return nil
}

func (s *storage) Save(metricType, metricName, metricValue string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch metricType {
	case "counter":
		metricCounterValue, err := metrics.ParseCounter(metricValue)

		if err != nil {
			return fmt.Errorf("%s: %w", metricValue, ErrInvalidMetricValue)
		}

		currentCounterValue, exist := s.metrics["counter"][metricName]

		if exist {
			s.metrics["counter"][metricName] = currentCounterValue.(metrics.Counter) + metricCounterValue
		} else {
			s.metrics["counter"][metricName] = metricCounterValue
		}

	case "gauge":
		metricGaugeValue, err := metrics.ParseGause(metricValue)

		if err != nil {
			return fmt.Errorf("%s: %w", metricValue, ErrInvalidMetricValue)
		}

		s.metrics["gauge"][metricName] = metricGaugeValue

	default:
		return fmt.Errorf("%s: %w", metricType, ErrInvalidMetricType)
	}

	return nil
}

func (s *storage) GetMetricValue(metricType, metricName string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	switch metricType {
	case "counter", "gauge":
		metricValue, exist := s.metrics[metricType][metricName]
		if exist {
			return metricValue, nil
		} else {
			return nil, fmt.Errorf("%s: %w", metricName, ErrNotFoundMetric)
		}
	default:
		return nil, fmt.Errorf("%s: %w", metricType, ErrInvalidMetricType)
	}
}

func (s *storage) GetMetrics() StorageMetrics {
	return s.metrics
}

func (s *storage) StoreMetrics(filePath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	fileObj, err := os.OpenFile(filePath, flag, 0644)
	defer func() {
		err := fileObj.Close()
		if err != nil {
			log.Printf("can't close file %s\n", filePath)
		}
	}()

	if err != nil {
		return fmt.Errorf("%w: can't open file %s", ErrStoreMetrics, filePath)
	}

	if err := json.NewEncoder(fileObj).Encode(s.GetMetrics()); err != nil {
		return fmt.Errorf("%w: cant't encode metrics", ErrStoreMetrics)
	}

	log.Printf("success save metrics to file %s\n", filePath)

	return nil
}

func (s *storage) LoadMetrics(filePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fileObj, err := os.Open(filePath)
	defer func() {
		err := fileObj.Close()
		if err != nil {
			log.Printf("can't close file %s\n", filePath)
		}
	}()

	if err != nil {
		return fmt.Errorf("%w: can't open file %s", ErrLoadMetrics, filePath)
	}

	if err := json.NewDecoder(fileObj).Decode(&s); err != nil {
		return fmt.Errorf("%w: cant't decode metrics", ErrLoadMetrics)
	}

	log.Printf("success load metrics from file %s\n", filePath)

	return nil
}

func NewStorage() Storage {
	return &storage{
		StorageMetrics{
			"counter": make(map[string]interface{}),
			"gauge":   make(map[string]interface{}),
		},
		sync.RWMutex{},
	}
}
