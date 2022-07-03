package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
)

var (
	ErrInvalidMetricValue = errors.New("invalid metric value")
	ErrInvalidMetricType  = errors.New("invalid metric type")
)

type (
	Gauge   float64
	Counter int64

	RuntimeMetrics struct {
		Alloc         Gauge
		BuckHashSys   Gauge
		Frees         Gauge
		GCCPUFraction Gauge
		GCSys         Gauge
		HeapAlloc     Gauge
		HeapIdle      Gauge
		HeapInuse     Gauge
		HeapObjects   Gauge
		HeapReleased  Gauge
		HeapSys       Gauge
		LastGC        Gauge
		Lookups       Gauge
		MCacheInuse   Gauge
		MCacheSys     Gauge
		MSpanInuse    Gauge
		MSpanSys      Gauge
		Mallocs       Gauge
		NextGC        Gauge
		NumForcedGC   Gauge
		NumGC         Gauge
		OtherSys      Gauge
		PauseTotalNs  Gauge
		StackInuse    Gauge
		StackSys      Gauge
		Sys           Gauge
		TotalAlloc    Gauge
		PollCount     Counter
		RandomValue   Gauge
	}

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
)

func (c Counter) ToInt64() int64 {
	return int64(c)
}

func (g Gauge) ToFloat64() float64 {
	return float64(g)
}

func (m *RuntimeMetrics) Collect() {
	memStats := new(runtime.MemStats)
	runtime.ReadMemStats(memStats)
	memStatsElements := reflect.ValueOf(memStats).Elem()
	metricsElements := reflect.ValueOf(m).Elem()

	for i := 0; i < memStatsElements.NumField(); i++ {
		for j := 0; j < metricsElements.NumField(); j++ {
			if memStatsElements.Type().Field(i).Name == metricsElements.Type().Field(j).Name {
				statValue := memStatsElements.Field(i).Interface()
				statValueConverted := reflect.ValueOf(statValue).Convert(metricsElements.Field(j).Type())
				metricsElements.Field(j).Set(statValueConverted)
			}
		}
	}

	m.PollCount++
	m.RandomValue = Gauge(rand.Float64())
}

func ParseCounter(s string) (Counter, error) {
	n, err := strconv.Atoi(s)

	if err != nil {
		return 0, fmt.Errorf("fnParseCounter: can't parse: %v", err)
	}

	return Counter(n), nil
}

func ParseGause(s string) (Gauge, error) {
	n, err := strconv.ParseFloat(s, 64)

	if err != nil {
		return 0, fmt.Errorf("fnParseGause: can't parse: %v", err)
	}

	return Gauge(n), nil
}

func NewMetric(metricID, metricType, metricValue string) (Metric, error) {
	var metric Metric

	metric.ID = metricID
	metric.MType = metricType

	switch metricType {
	case "counter":
		mValue, err := strconv.ParseInt(metricValue, 10, 64)

		if err != nil {
			return metric, fmt.Errorf("Metric_NewMetric error: %w", ErrInvalidMetricValue)
		}
		metric.Delta = &mValue
		return metric, nil

	case "gauge":
		mValue, err := strconv.ParseFloat(metricValue, 64)

		if err != nil {
			return metric, fmt.Errorf("Metric_NewMetric error: %w", ErrInvalidMetricValue)
		}
		metric.Value = &mValue

		return metric, nil

	default:
		return metric, fmt.Errorf("Metric_NewMetric error: %w", ErrInvalidMetricType)
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
			return fmt.Errorf("Metric_Valid error: %w", ErrInvalidMetricValue)
		}
	case "gauge":
		if m.Value == nil {
			return fmt.Errorf("Metric_Valid error: %w", ErrInvalidMetricValue)
		}
	default:
		return fmt.Errorf("Metric_Valid error: %w", ErrInvalidMetricType)
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
		return nil, fmt.Errorf("Metric_GetMetrics error: %w", ErrInvalidMetricType)
	}
}

func (m *Metric) CalcHash(key string) (string, error) {
	var msg string
	switch m.MType {
	case "counter":
		msg = fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)
	case "gauge":
		msg = fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)
	default:
		return "", fmt.Errorf("Metric_CalcHash error: %w", ErrInvalidMetricType)
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg))
	hash := h.Sum(nil)
	return fmt.Sprintf("%x", hash), nil
}
