package collector

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"sync"

	"github.com/sreway/yametrics/internal/metrics"

	"github.com/shirou/gopsutil/v3/mem"
)

type (
	Gauge   float64
	Counter int64

	Metrics struct {
		Alloc           Gauge
		BuckHashSys     Gauge
		Frees           Gauge
		GCCPUFraction   Gauge
		GCSys           Gauge
		HeapAlloc       Gauge
		HeapIdle        Gauge
		HeapInuse       Gauge
		HeapObjects     Gauge
		HeapReleased    Gauge
		HeapSys         Gauge
		LastGC          Gauge
		Lookups         Gauge
		MCacheInuse     Gauge
		MCacheSys       Gauge
		MSpanInuse      Gauge
		MSpanSys        Gauge
		Mallocs         Gauge
		NextGC          Gauge
		NumForcedGC     Gauge
		NumGC           Gauge
		OtherSys        Gauge
		PauseTotalNs    Gauge
		StackInuse      Gauge
		StackSys        Gauge
		Sys             Gauge
		TotalAlloc      Gauge
		TotalMemory     Gauge
		FreeMemory      Gauge
		CPUutilization1 Gauge
		PollCount       Counter
		RandomValue     Gauge
		mu              sync.RWMutex
	}
)

func (c Counter) ToInt64() int64 {
	return int64(c)
}

func (g Gauge) ToFloat64() float64 {
	return float64(g)
}

func (m *Metrics) CollectRuntimeMetrics() {
	rtm := new(runtime.MemStats)
	runtime.ReadMemStats(rtm)
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Alloc = Gauge(rtm.Alloc)
	m.BuckHashSys = Gauge(rtm.BuckHashSys)
	m.Frees = Gauge(rtm.Frees)
	m.GCCPUFraction = Gauge(rtm.GCCPUFraction)
	m.GCSys = Gauge(rtm.GCSys)
	m.HeapAlloc = Gauge(rtm.HeapAlloc)
	m.HeapIdle = Gauge(rtm.HeapIdle)
	m.HeapInuse = Gauge(rtm.HeapInuse)
	m.HeapObjects = Gauge(rtm.HeapObjects)
	m.HeapReleased = Gauge(rtm.HeapReleased)
	m.HeapSys = Gauge(rtm.HeapSys)
	m.LastGC = Gauge(rtm.LastGC)
	m.Lookups = Gauge(rtm.Lookups)
	m.MCacheInuse = Gauge(rtm.MCacheInuse)
	m.MCacheSys = Gauge(rtm.MCacheSys)
	m.MSpanInuse = Gauge(rtm.MSpanInuse)
	m.MSpanSys = Gauge(rtm.MSpanSys)
	m.Mallocs = Gauge(rtm.Mallocs)
	m.NextGC = Gauge(rtm.NextGC)
	m.NumForcedGC = Gauge(rtm.NumForcedGC)
	m.NumGC = Gauge(rtm.NumGC)
	m.OtherSys = Gauge(rtm.OtherSys)
	m.PauseTotalNs = Gauge(rtm.PauseTotalNs)
	m.StackInuse = Gauge(rtm.StackInuse)
	m.StackSys = Gauge(rtm.StackSys)
	m.Sys = Gauge(rtm.Sys)
	m.TotalAlloc = Gauge(rtm.TotalAlloc)
	m.PollCount++
	m.RandomValue = Gauge(rand.Float64())
}

func (m *Metrics) CollectMemmoryMetrics() {
	memStats, _ := mem.VirtualMemory()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalMemory = Gauge(memStats.Total)
	m.FreeMemory = Gauge(memStats.Free)
}

func (m *Metrics) SetCPUutilization(cpuUtilization Gauge) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CPUutilization1 = cpuUtilization
}

func (m *Metrics) ExposeMetrics() []metrics.Metric {
	m.mu.Lock()
	defer m.mu.Unlock()

	metricsElements := reflect.ValueOf(m).Elem()
	exposeMetrics := make([]metrics.Metric, 0, metricsElements.NumField())

	for i := 0; i < metricsElements.NumField(); i++ {
		exposeMetric := metrics.Metric{
			ID: metricsElements.Type().Field(i).Name,
		}
		switch metricsElements.Field(i).Type().Name() {
		case "Gauge":
			metricValue := metricsElements.Field(i).Interface().(Gauge).ToFloat64()
			exposeMetric.MType = "gauge"
			exposeMetric.Value = &metricValue
		case "Counter":
			metricValue := metricsElements.Field(i).Interface().(Counter).ToInt64()
			exposeMetric.MType = "counter"
			exposeMetric.Delta = &metricValue
		default:
			continue
		}

		exposeMetrics = append(exposeMetrics, exposeMetric)
	}

	return exposeMetrics
}

func (m *Metrics) ClearPollCounter() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PollCount = 0
}

func ParseCounter(s string) (Counter, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("fnParseCounter: can't parse: %w", err)
	}

	return Counter(n), nil
}

func ParseGause(s string) (Gauge, error) {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("fnParseGause: can't parse: %w", err)
	}

	return Gauge(n), nil
}
