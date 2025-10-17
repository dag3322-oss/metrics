package main

import (
	"flag"
	"log"
	"net"

	"github.com/dag3322-oss/metrics/internal/application"
)

func main() {
	var host = flag.String("a", "localhost:8080", "host:port")
	flag.Parse()
	if !flag.Parsed() {
		flag.PrintDefaults()
		return
	}
	_, _, err := net.SplitHostPort(*host)
	if err != nil {
		log.Printf("%s", err.Error())
		flag.PrintDefaults()
		return
	}
	var server = application.Server{}
	server.SetHost(*host)
	server.Run()
}
