// Package agent implements and describes an agent for collecting and sending them to the server
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

// Agent describes the implementation of collector
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

// CollectRuntimeMetrics implements collects runtime metrics
func (a *agent) CollectRuntimeMetrics(ctx context.Context, wg *sync.WaitGroup) {
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

// CollectUtilMetrics implements collects memory and cpu metrics
func (a *agent) CollectUtilMetrics(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	cpuUsage := make(chan collector.Gauge)
	stopCh := make(chan struct{})

	go CollectCPUInfo(ctx, wg, cpuUsage, stopCh)

	for {
		select {
		case cpuData := <-cpuUsage:
			a.collector.CollectUtilMetrics(cpuData)
		case <-ctx.Done():
			close(stopCh)
			return
		}
	}
}

// Send implements periodic sending of metrics to the server
func (a *agent) Send(ctx context.Context, wg *sync.WaitGroup) {
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

// Start implements starting/stopping the agent and running periodic tasks
func (a *agent) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	systemSignals := make(chan os.Signal, 1)
	signal.Notify(systemSignals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	exitChan := make(chan int)
	wg := new(sync.WaitGroup)
	wg.Add(4)
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

// NewAgent implements agent initialization
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

// SendToSever implements sending metrics to the server
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
		return fmt.Errorf("failed encode metric: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, a.Config.metricEndpoint, &body)
	if err != nil {
		return fmt.Errorf("failed create request: %w", err)
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := a.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed send request: %w", err)
	}

	err = response.Body.Close()
	if err != nil {
		return fmt.Errorf("failed close response body: %w", err)
	}

	return nil
}

func getCPUInfo() collector.Gauge {
	percent, _ := cpu.Percent(10*time.Second, false)
	return collector.Gauge(percent[0])
}

// CollectCPUInfo implements the collection of CPU info and sending them to the data channel
func CollectCPUInfo(ctx context.Context, wg *sync.WaitGroup, dataCh chan collector.Gauge, stopCh chan struct{}) {
	defer wg.Done()

	for {
		// try exit early
		select {
		case <-stopCh:
			return
		default:
		}

		select {
		case <-ctx.Done():
			return
		case <-stopCh:
			return
		case dataCh <- getCPUInfo():
		}
	}
}
