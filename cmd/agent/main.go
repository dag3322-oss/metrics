package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime/metrics"
	"syscall"
	"time"

	repository "github.com/dag3322-oss/metrics/internal/repository"
)

func main() {
	file, err := os.OpenFile("agent.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	log.SetOutput(file)

	var repo = repository.NewMemRepository()

	Collect(time.Now(), repo)
	var collectTimer = time.NewTicker(time.Second * 2)
	go CollectEvent(collectTimer, repo)

	var httpc = http.Client{Timeout: time.Duration(1) * time.Second}
	var sendTimer = time.NewTicker(time.Second * 10)
	go SendEvent(sendTimer, repo, httpc)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	collectTimer.Stop()
	//os.Exit(1)
}

func GetSamples() []metrics.Sample {
	var s = []metrics.Sample{
		{Name: "Alloc"},
		{Name: "BuckHashSys"},
		{Name: "Frees"},
		{Name: "GCCPUFraction"},
		{Name: "GCSys"},
		{Name: "HeapAlloc"},
		{Name: "HeapIdle"},
		{Name: "HeapInuse"},
		{Name: "HeapObjects"},
		{Name: "HeapReleased"},
		{Name: "HeapSys"},
		{Name: "LastGC"},
		{Name: "Lookups"},
		{Name: "MCacheInuse"},
		{Name: "MCacheSys"},
		{Name: "MSpanInuse"},
		{Name: "MSpanSys"},
		{Name: "Mallocs"},
		{Name: "NextGC"},
		{Name: "NumForcedGC"},
		{Name: "NumGC"},
		{Name: "OtherSys"},
		{Name: "PauseTotalNs"},
		{Name: "StackInuse"},
		{Name: "StackSys"},
		{Name: "Sys"},
		{Name: "TotalAlloc"},
	}
	return s
}

func CollectEvent(tick *time.Ticker, repo repository.Repository) {
	for t := range tick.C {
		Collect(t, repo)
	}
}

func Collect(t time.Time, repo repository.Repository) error {
	var desc string
	for _, d := range metrics.All() {
		desc = desc + "\n" + d.Name + " | " + d.Description
	}
	log.Printf("%s", desc)
	var samples = GetSamples()
	metrics.Read(samples)
	for _, s := range samples {
		if s.Value.Kind() == metrics.KindFloat64 {
			var err = repo.UpdateMetric(s.Name, s.Value.Float64())
			if err != nil {
				return err
			}
		}
	}
	repo.UpdateMetric("RandomValue", rand.Float64())
	repo.UpdateMetric("PollCount", int64(1))
	return nil
}

func SendEvent(tick *time.Ticker, repo repository.Repository, httpc http.Client) {
	for range tick.C {
		Send(repo, httpc)
	}
}

func Send(repo repository.Repository, httpc http.Client) error {
	var metricType string
	for k, v := range repo.GetAll() {
		switch v.(type) {
		case float64:
			metricType = "gauge"
		case int64:
			metricType = "counter"
		default:
			return fmt.Errorf("invalid metric type %s", reflect.TypeOf(v).Name())
		}
		resp, err := httpc.Get(fmt.Sprintf("https://%s/update/%s/%s/%+v", metricType, k, v))
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HPPT status code = %d", resp.StatusCode)
		}
		defer resp.Body.Close()
	}
	log.Printf("metrics sended")
	return nil
}
