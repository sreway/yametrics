package server

import (
	"errors"
	"fmt"
	"github.com/sreway/yametrics/internal/metrics"
	"sync"
)

var (
	ErrInvalidMetricValue = errors.New("invalid metric value")
	ErrInvalidMetricType  = errors.New("invalid metric type")
	ErrNotFoundMetric     = errors.New("not found metric")
)

type Storage interface {
	Save(metricType, metricName, metricValue string) error
	GetMetricValue(metricType, metricName string) (interface{}, error)
	GetMetrics() map[string]map[string]interface{}
}

type storage struct {
	metrics map[string]map[string]interface{}
	mu      sync.RWMutex
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

func (s *storage) GetMetrics() map[string]map[string]interface{} {
	return s.metrics
}

func NewStorage() Storage {
	return &storage{
		map[string]map[string]interface{}{
			"counter": make(map[string]interface{}),
			"gauge":   make(map[string]interface{}),
		},
		sync.RWMutex{},
	}
}
