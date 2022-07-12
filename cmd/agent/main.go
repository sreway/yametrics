package main

import (
	"flag"
	"github.com/sreway/yametrics/internal/agent"
	"log"
)

func init() {
	flag.StringVar(&agent.ServerAddressDefault, "a", agent.ServerAddressDefault,
		"server address: host:port")
	flag.DurationVar(&agent.ReportIntervalDefault, "r", agent.ReportIntervalDefault, "report interval")
	flag.DurationVar(&agent.PollIntervalDefault, "p", agent.PollIntervalDefault, "poll interval")
	flag.StringVar(&agent.KeyDefault, "k", agent.KeyDefault, "encrypt key")
	flag.Parse()
}

func main() {
	cli, err := agent.NewAgent()
	if err != nil {
		log.Fatalln(err)
	}


	cli.Start()
}
