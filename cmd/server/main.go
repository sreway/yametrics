package main

import (
	"github.com/sreway/yametrics/internal/server"
	"log"
)

func main() {
	serv, err := server.NewServer()
	if err != nil {
		log.Fatalln(err)
	}
	serv.Start()
}
