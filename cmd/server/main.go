package main

import (
	"github.com/sreway/yametrics/internal/server"
)

const (
	serverAddr = "127.0.0.1"
	serverPort = "8080"
)

func main() {
	servConfig := server.NewServerConfig(serverAddr, serverPort)
	serv := server.NewServer(servConfig)
	serv.Start()
}
