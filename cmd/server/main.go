package main

import (
	"flag"
	"log"

	"github.com/sreway/yametrics/internal/server"
)

func main() {
	flag.StringVar(&server.AddressDefault, "a", server.AddressDefault, "address: host:port")
	flag.DurationVar(&server.StoreIntervalDefault, "i", server.StoreIntervalDefault, "store interval")
	flag.BoolVar(&server.RestoreDefault, "r", server.RestoreDefault, "restoring metrics at startup")
	flag.StringVar(&server.StoreFileDefault, "f", server.StoreFileDefault, "store file")
	flag.StringVar(&server.KeyDefault, "k", server.KeyDefault, "encrypt key")
	flag.StringVar(&server.DsnDefault, "d", server.DsnDefault, "PosgreSQL data source name")
	flag.Parse()

	serv, err := server.NewServer()
	if err != nil {
		log.Fatalln(err)
	}
	serv.Start()
}
