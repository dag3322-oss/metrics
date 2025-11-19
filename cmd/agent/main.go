package main

import (
	"github.com/dag3322-oss/metrics/internal/application"
)

func main() {
	var agent = application.Agent{}
	agent.Run(nil)
}
