package main

import (
	"log"
	"net/http"
	"os"

	updater "github.com/dag3322-oss/metrics/internal/handler"
	metrics "github.com/dag3322-oss/metrics/internal/repository"
)

func main() {
	file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	log.SetOutput(file)

	var repo = metrics.NewMemRepository()

	var hu = updater.NewMetricUpdateHandler(repo)
	http.HandleFunc(`/update/`, hu.Handle)

	var hl = updater.NewMetricListHandler(repo)
	http.HandleFunc(`/list/`, hl.Handle)

	err = http.ListenAndServe(`:8080`, nil)
	if err != nil {
		log.Fatal("Failed to service:", err)
	}
}
