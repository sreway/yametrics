package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrInvalidMetricValue = errors.New("invalid metric value")
	ErrInvalidMetricType  = errors.New("invalid metric type")
)

type (
	Metric struct {
		ID    string   `json:"id" db:"name"`
		MType string   `json:"type" db:"type"`
		Delta *int64   `json:"delta,omitempty" db:"delta"`
		Value *float64 `json:"value,omitempty" db:"value"`
		Hash  string   `json:"hash,omitempty"`
	}

	Metrics struct {
		Counter map[string]Metric `json:"counter"`
		Gauge   map[string]Metric `json:"gauge"`
	}

	ErrMetric struct {
		MetricError error
		MetricType  string
		MetricID    string
	}
	MetricValue interface {
		float64 | int64
	}
)

func NewMetric(metricID, metricType, metricValue string) (Metric, error) {
	var metric Metric

	metric.ID = metricID
	metric.MType = metricType

	switch metricType {
	case "counter":
		mValue, err := strconv.ParseInt(metricValue, 10, 64)

		if err != nil {
			return metric, fmt.Errorf("NewMetric: %w",
				NewMetricError(metricType, metricID, ErrInvalidMetricValue))
		}
		metric.Delta = &mValue
		return metric, nil

	case "gauge":
		mValue, err := strconv.ParseFloat(metricValue, 64)

		if err != nil {
			return metric, fmt.Errorf("NewMetric: %w",
				NewMetricError(metricType, metricID, ErrInvalidMetricValue))
		}
		metric.Value = &mValue

		return metric, nil

	default:
		return metric, fmt.Errorf("NewMetric: %w",
			NewMetricError(metricType, metricID, ErrInvalidMetricType))
	}
}

func (m Metric) IsCounter() bool {
	return m.MType == "counter"
}

func (m Metric) GetStrValue() string {
	switch m.MType {
	case "counter":
		return fmt.Sprintf("%v", *m.Delta)
	case "gauge":
		return fmt.Sprintf("%v", *m.Value)
	default:
		return ""
	}
}

func (m *Metric) Float64Value() float64 {
	if m.Value == nil {
		return 0
	}
	return *m.Value
}

func (m *Metric) Float64Pointer() *float64 {
	return m.Value
}

func (m *Metric) Int64Value() int64 {
	if m.Delta == nil {
		return 0
	}
	return *m.Delta
}

func (m *Metric) Int64Pointer() *int64 {
	return m.Delta
}

func (m *Metric) SetFloat64(f float64) {
	m.Value = &f
}

func (m *Metric) SetInt64(i int64) {
	m.Delta = &i
}

func (m *Metric) Valid() error {
	switch m.MType {
	case "counter":
		if m.Delta == nil {
			return fmt.Errorf("Metric_Valid: %w",
				NewMetricError(m.MType, m.ID, ErrInvalidMetricValue))
		}
	case "gauge":
		if m.Value == nil {
			return fmt.Errorf("Metric_Valid: %w",
				NewMetricError(m.MType, m.ID, ErrInvalidMetricValue))
		}
	default:
		return fmt.Errorf("Metric_Valid: %w",
			NewMetricError(m.MType, m.ID, ErrInvalidMetricType))
	}
	return nil
}

func (m *Metrics) GetMetrics(metricsType string) (map[string]Metric, error) {
	switch metricsType {
	case "counter":
		return m.Counter, nil
	case "gauge":
		return m.Gauge, nil
	default:
		return nil, fmt.Errorf("Metrics_GetMetrics: %w",
			NewMetricError(metricsType, "", ErrInvalidMetricType))
	}
}

func calcHash[T MetricValue](key, metricID, metricType string, metricValue T) string {
	var msg string

	intValue, ok := any(metricValue).(int64)
	if ok {
		msg = fmt.Sprintf("%s:%s:%d", metricID, metricType, intValue)
	}

	floatValue, ok := any(metricValue).(float64)
	if ok {
		msg = fmt.Sprintf("%s:%s:%f", metricID, metricType, floatValue)
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg))
	hash := h.Sum(nil)
	return fmt.Sprintf("%x", hash)
}

func (m *Metric) CalcHash(key string) string {
	if m.IsCounter() {
		return calcHash[int64](key, m.ID, m.MType, *m.Delta)
	}

	return calcHash[float64](key, m.ID, m.MType, *m.Value)
}

func (e *ErrMetric) Error() string {
	return fmt.Sprintf("[%s][%s] error %s", e.MetricType, e.MetricID, e.MetricError)
}

func NewMetricError(metricType, metricID string, err error) error {
	return &ErrMetric{
		MetricType:  metricType,
		MetricID:    metricID,
		MetricError: err,
	}
}
