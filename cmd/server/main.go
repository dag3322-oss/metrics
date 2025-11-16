package main

import (
	"github.com/dag3322-oss/metrics/internal/application"
)

func main() {
	var server = application.Server{}
	server.Run(nil)
}
