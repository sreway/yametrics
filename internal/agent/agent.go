package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/sreway/yametrics/internal/collector"
	"github.com/sreway/yametrics/internal/metrics"
)

type Agent interface {
	Start()
	CollectRuntimeMetrics(ctx context.Context, wg *sync.WaitGroup)
	Send(ctx context.Context, wg *sync.WaitGroup)
}

type agent struct {
	collector  collector.Collector
	httpClient http.Client
	Config     *agentConfig
}

func (a *agent) CollectRuntimeMetrics(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	tick := time.NewTicker(a.Config.PollInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			a.collector.CollectRuntimeMetrics()

		case <-ctx.Done():
			return
		}
	}
}

func (a *agent) CollectUtilMetrics(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	cpuUsage := make(chan collector.Gauge)

	go CollectCPUInfo(ctx, wg, cpuUsage)

	for {
		select {
		case cpuData := <-cpuUsage:
			a.collector.CollectUtilMetrics(cpuData)
		case <-ctx.Done():
			close(cpuUsage)
			return
		}
	}
}

func (a *agent) Send(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	tick := time.NewTicker(a.Config.ReportInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			exposeMetrics := a.collector.ExposeMetrics()
			err := a.SendToSever(exposeMetrics, a.Config.Key != "")

			if err != nil {
				log.Printf("agent send error: %v", err)
			} else {
				a.collector.ClearPollCounter()
			}

		case <-ctx.Done():
			return
		}
	}
}

func (a *agent) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	systemSignals := make(chan os.Signal)
	signal.Notify(systemSignals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	exitChan := make(chan int)
	wg := new(sync.WaitGroup)

	go a.CollectRuntimeMetrics(ctx, wg)
	go a.CollectUtilMetrics(ctx, wg)

	go a.Send(ctx, wg)
	go func() {
		for {
			s := <-systemSignals
			switch s {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				log.Println("signal triggered.")
				exitChan <- 0
			default:
				log.Println("unknown signal.")
				exitChan <- 1
			}
		}
	}()

	exitCode := <-exitChan
	cancel()
	wg.Wait()
	os.Exit(exitCode)
}

func NewAgent(opts ...OptionAgent) (Agent, error) {
	agentCfg, err := newAgentConfig()
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		err := opt(agentCfg)
		if err != nil {
			return nil, err
		}
	}

	return &agent{
		collector:  collector.NewCollector(),
		Config:     agentCfg,
		httpClient: http.Client{},
	}, nil
}

func (a *agent) SendToSever(m []metrics.Metric, withHash bool) error {
	var body bytes.Buffer

	if len(m) == 0 {
		return nil
	}

	for index, metric := range m {
		if withHash {
			sign := metric.CalcHash(a.Config.Key)
			m[index].Hash = sign
		}
	}

	if err := json.NewEncoder(&body).Encode(&m); err != nil {
		return fmt.Errorf("failed encode metric: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, a.Config.metricEndpoint, &body)
	if err != nil {
		return fmt.Errorf("failed create request: %v", err)
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := a.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed send request: %v", err)
	}

	err = response.Body.Close()
	if err != nil {
		return fmt.Errorf("failed close response body: %v", err)
	}

	return nil
}

func getCPUInfo() collector.Gauge {
	percent, _ := cpu.Percent(10*time.Second, false)
	return collector.Gauge(percent[0])
}

func CollectCPUInfo(ctx context.Context, wg *sync.WaitGroup, cpuUsage chan collector.Gauge) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			cpuUsage <- getCPUInfo()
		}
	}
}
