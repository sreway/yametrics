package main

import (
	"github.com/sreway/yametrics/internal/agent"
	"time"
)

const (
	pollInterval      = 2 * time.Second
	reportInterval    = 10 * time.Second
	serverAddr        = "127.0.0.1"
	serverPort        = "8080"
	serverScheme      = "http"
	serverContentType = "text/plain"
)

func main() {
	cliConfig := agent.NewAgentConfig(pollInterval, reportInterval, serverAddr,
		serverPort, serverScheme, serverContentType)
	cli := agent.NewAgent(cliConfig)
	cli.Start()
}
