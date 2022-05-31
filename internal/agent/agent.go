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
	Collect(ctx context.Context, wg *sync.WaitGroup)
	Send(ctx context.Context, wg *sync.WaitGroup)
}

type agent struct {
	collector  Collector
	httpClient http.Client
	Config     *agentConfig
}

func (a *agent) Collect(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	tick := time.NewTicker(a.Config.pollInterval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			a.collector.CollectMetrics()
		case <-ctx.Done():
			return
		}
	}
}

func (a *agent) Send(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
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

	go a.Collect(ctx, wg)
	go a.Send(ctx, wg)
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
	cancel()
	wg.Wait()
	os.Exit(exitCode)
}

func NewAgent(opts ...OptionAgent) (Agent, error) {
	agentCfg := newAgentConfig()

	for _, opt := range opts {
		err := opt(agentCfg)
		if err != nil {
			return nil, err
		}
	}

	return &agent{
		collector:  NewCollector(),
		Config:     agentCfg,
		httpClient: http.Client{},
	}, nil
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

		request.Header.Add("Content-Type", "text/plain")

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
