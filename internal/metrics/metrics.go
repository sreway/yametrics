package metrics

import (
	"math/rand"
	"reflect"
	"runtime"
)

type Gauge float64
type Counter int64

type Metrics struct {
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

func (m *Metrics) Collect() {
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
