// Package metrics implements and describes a common type of metric for agent and server
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

const (
	CounterStrName = "counter"
	GaugeStrName   = "gauge"
)

type (
	// Metric defines the type of metric to send to the server
	Metric struct {
		ID    string   `json:"id" db:"name"`
		MType string   `json:"type" db:"type"`
		Delta *int64   `json:"delta,omitempty" db:"delta"`
		Value *float64 `json:"value,omitempty" db:"value"`
		Hash  string   `json:"hash,omitempty"`
	}
	// Metrics defines the type for counter and gauge metrics
	Metrics struct {
		Counter map[string]Metric `json:"counter"`
		Gauge   map[string]Metric `json:"gauge"`
	}
	// ErrMetric defines metric error
	ErrMetric struct {
		MetricError error
		MetricType  string
		MetricID    string
	}
	// MetricValue defines metric value
	MetricValue interface {
		float64 | int64
	}
)

// NewMetric implements creating a metric
func NewMetric(metricID, metricType, metricValue string) (Metric, error) {
	var metric Metric

	metric.ID = metricID
	metric.MType = metricType

	switch metricType {
	case CounterStrName:
		mValue, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return metric, fmt.Errorf("NewMetric: %w",
				NewMetricError(metricType, metricID, ErrInvalidMetricValue))
		}
		metric.Delta = &mValue
		return metric, nil

	case GaugeStrName:
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

// IsCounter implements the check that the metric is a counter
func (m Metric) IsCounter() bool {
	return m.MType == "counter"
}

// GetStrValue implements getting the string value of the metric
func (m Metric) GetStrValue() string {
	switch m.MType {
	case CounterStrName:
		return fmt.Sprintf("%v", *m.Delta)
	case GaugeStrName:
		return fmt.Sprintf("%v", *m.Value)
	default:
		return ""
	}
}

// Float64Value implements getting the float64 value of the metric
func (m *Metric) Float64Value() float64 {
	if m.Value == nil {
		return 0
	}
	return *m.Value
}

// Float64Pointer implements getting the float64 pointer of the metric
func (m *Metric) Float64Pointer() *float64 {
	return m.Value
}

// Int64Value implements getting the int64 value of the metric
func (m *Metric) Int64Value() int64 {
	if m.Delta == nil {
		return 0
	}
	return *m.Delta
}

// Int64Pointer implements getting the int64 pointer of the metric
func (m *Metric) Int64Pointer() *int64 {
	return m.Delta
}

// SetFloat64 implements set float64 value of the metric
func (m *Metric) SetFloat64(f float64) {
	m.Value = &f
}

// SetInt64 implements set int64 value of the metric
func (m *Metric) SetInt64(i int64) {
	m.Delta = &i
}

// Valid implements validation metric
func (m *Metric) Valid() error {
	switch m.MType {
	case CounterStrName:
		if m.Delta == nil {
			return fmt.Errorf("Metric_Valid: %w",
				NewMetricError(m.MType, m.ID, ErrInvalidMetricValue))
		}
	case GaugeStrName:
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

// GetMetrics implements getting metrics depending on the type
func (m *Metrics) GetMetrics(metricsType string) (map[string]Metric, error) {
	switch metricsType {
	case CounterStrName:
		return m.Counter, nil
	case GaugeStrName:
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

// CalcHash implements the calculation of the metric hash
func (m *Metric) CalcHash(key string) string {
	if m.IsCounter() {
		return calcHash[int64](key, m.ID, m.MType, *m.Delta)
	}

	return calcHash[float64](key, m.ID, m.MType, *m.Value)
}

// ErrMetric implements metric error stringer
func (e *ErrMetric) Error() string {
	return fmt.Sprintf("[%s][%s] error %s", e.MetricType, e.MetricID, e.MetricError)
}

// NewMetricError implements the creation of a metric error
func NewMetricError(metricType, metricID string, err error) error {
	return &ErrMetric{
		MetricType:  metricType,
		MetricID:    metricID,
		MetricError: err,
	}
}
