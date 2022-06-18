package main

import (
	"github.com/sreway/yametrics/internal/agent"
	"log"
	"os"
)

func init() {
	os.Environ()
}

func main() {
	cli, err := agent.NewAgent(agent.WithReportInterval("5s"))
	if err != nil {
		log.Fatalln(err)
	}
	cli.Start()
}
