package main

import (
	"flag"
	"net"

	"github.com/rs/zerolog/log"

	"github.com/dag3322-oss/metrics/internal/application"
)

func main() {
	var host = flag.String("a", "localhost:8080", "host:port")
	var reportInterval = flag.Int64("r", 10, "send interval(seconds)")
	var pollInterval = flag.Int64("p", 2, "poll interval(seconds)")
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

	var agent = application.Agent{}

	agent.SetHost(*host)
	agent.SetReportInterval(*reportInterval)
	agent.SetPollInterval(*pollInterval)
	log.Printf("start agent host=%s,reportInterval=%d,pollInterval=%d", *host, *reportInterval, *pollInterval)
	agent.Run()
}
