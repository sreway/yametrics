package agent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Agent interface {
	Start()
	Stop()
	Collect()
	Send()
}

type agentConfig struct {
	pollInterval      time.Duration
	reportInterval    time.Duration
	serverAddr        string
	serverPort        string
	serverScheme      string
	serverContentType string
	serverURL         string
}

func NewAgentConfig(poolInterval, reportInterval time.Duration, serverAddr string,
	serverPort string, serverScheme string, serverContetType string) *agentConfig {
	return &agentConfig{
		pollInterval:      poolInterval,
		reportInterval:    reportInterval,
		serverAddr:        serverAddr,
		serverPort:        serverPort,
		serverScheme:      serverScheme,
		serverContentType: serverContetType,
		serverURL:         fmt.Sprintf("%s://%s:%s", serverScheme, serverAddr, serverPort),
	}
}

type agent struct {
	collector  Collector
	ctx        context.Context
	stopFunc   context.CancelFunc
	wg         *sync.WaitGroup
	httpClient http.Client
	Config     *agentConfig
}

func (a *agent) Collect() {
	a.wg.Add(1)
	defer a.wg.Done()
	tick := time.NewTicker(a.Config.pollInterval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			a.collector.CollectMetrics()
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *agent) Send() {
	a.wg.Add(1)
	defer a.wg.Done()
	tick := time.NewTicker(a.Config.reportInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			exposeMetrics := a.collector.ExposeMetrics()
			err := a.SendToSever(exposeMetrics)

			if err != nil {
				log.Printf("agent send error: %v", err)
			} else {
				a.collector.ClearPollCounter()
			}

		case <-a.ctx.Done():
			return
		}
	}
}

func (a *agent) Start() {
	systemSignals := make(chan os.Signal)
	signal.Notify(systemSignals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	exitChan := make(chan int)
	go a.Collect()
	go a.Send()
	go func() {
		for {
			s := <-systemSignals
			switch s {
			case syscall.SIGINT:
				log.Println("signal interrupt triggered.")
				exitChan <- 0
			case syscall.SIGTERM:
				log.Println("signal terminate triggered.")
				exitChan <- 0
			case syscall.SIGQUIT:
				log.Println("signal quit triggered.")
				exitChan <- 0
			default:
				log.Println("unknown signal.")
				exitChan <- 1
			}
		}
	}()
	exitCode := <-exitChan
	a.Stop()
	os.Exit(exitCode)
}

func (a *agent) Stop() {
	a.stopFunc()
	a.wg.Wait()
}

func NewAgent(config *agentConfig) Agent {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	return &agent{
		collector:  NewCollector(),
		ctx:        ctx,
		stopFunc:   cancel,
		wg:         wg,
		Config:     config,
		httpClient: http.Client{},
	}
}

func (a *agent) SendToSever(metrics []ExposeMetric) error {
	for _, metric := range metrics {
		metricURI := fmt.Sprintf("update/%s/%s/%v", strings.ToLower(metric.Type),
			metric.ID, metric.Value)
		endpoint := fmt.Sprintf("%s/%s", a.Config.serverURL, metricURI)
		request, err := http.NewRequest(http.MethodPost, endpoint, nil)

		if err != nil {
			return fmt.Errorf("failed create request: %v", err)
		}

		request.Header.Add("Content-Type", a.Config.serverContentType)

		response, err := a.httpClient.Do(request)

		if err != nil {
			return fmt.Errorf("failed send request: %v", err)
		}

		err = response.Body.Close()

		if err != nil {
			return fmt.Errorf("failed close response body: %v", err)
		}
	}
	return nil
}
