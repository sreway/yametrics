package collector

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"

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
	}
)

func (c Counter) ToInt64() int64 {
	return int64(c)
}

func (g Gauge) ToFloat64() float64 {
	return float64(g)
}

// rename to runtime
func (m *Metrics) CollectRuntimeMetrics() {
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

func (m *Metrics) CollectMemmoryMetrics() {
	memStats, _ := mem.VirtualMemory()
	m.TotalMemory = Gauge(memStats.Total)
	m.FreeMemory = Gauge(memStats.Free)
}

//	m.CPUutilization1 = Gauge(getCPUutilization(10 * time.Second))
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
