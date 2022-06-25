package main

import (
	"flag"
	"github.com/sreway/yametrics/internal/server"
	"log"
)

func main() {
	flag.StringVar(&server.AddressDefault, "a", server.AddressDefault, "address: host:port")
	flag.DurationVar(&server.StoreIntervalDefault, "i", server.StoreIntervalDefault, "store interval")
	flag.BoolVar(&server.RestoreDefault, "r", server.RestoreDefault, "restoring metrics at startup")
	flag.StringVar(&server.StoreFileDefault, "f", server.StoreFileDefault, "store file")
	flag.Parse()

	serv, err := server.NewServer()
	if err != nil {
		log.Fatalln(err)
	}
	serv.Start()
}
