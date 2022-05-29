package server

import (
	"errors"
	"github.com/sreway/yametrics/internal/metrics"
	"sync"
)

var (
	ErrInvalidMetricValue = errors.New("invalid metric value")
	ErrInvalidMetricType  = errors.New("invalid metric type")
)

type Storage interface {
	Save(metricType, metricName, metricValue string) error
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
			return ErrInvalidMetricValue
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
			return ErrInvalidMetricValue
		}

		s.metrics["gauge"][metricName] = metricGaugeValue

	default:
		return ErrInvalidMetricType
	}

	return nil
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
