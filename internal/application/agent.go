package application

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"runtime/metrics"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	repository "github.com/dag3322-oss/metrics/internal/repository"
)

type Agent struct{}

func (a Agent) Run() {
	/* 	file, err := os.OpenFile("agent.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	   	if err != nil {
	   		log.Fatal("Failed to open log file:", err)
	   	}
	   	log.SetOutput(file)
	*/
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
	sendTimer.Stop()
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
	/* 	var desc string
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
	*/
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	repo.UpdateMetric("Alloc", m.Alloc)
	repo.UpdateMetric("BuckHashSys", m.BuckHashSys)
	repo.UpdateMetric("Frees", m.Frees)
	repo.UpdateMetric("GCCPUFraction", m.GCCPUFraction)
	repo.UpdateMetric("GCSys", m.GCSys)
	repo.UpdateMetric("HeapAlloc", m.HeapAlloc)
	repo.UpdateMetric("HeapIdle", m.HeapIdle)
	repo.UpdateMetric("HeapInuse", m.HeapInuse)
	repo.UpdateMetric("HeapObjects", m.HeapObjects)
	repo.UpdateMetric("HeapReleased", m.HeapReleased)
	repo.UpdateMetric("HeapSys", m.HeapSys)
	repo.UpdateMetric("LastGC", m.LastGC)
	repo.UpdateMetric("Lookups", m.Lookups)
	repo.UpdateMetric("MCacheInuse", m.MCacheInuse)
	repo.UpdateMetric("MCacheSys", m.MCacheSys)
	repo.UpdateMetric("MSpanInuse", m.MSpanInuse)
	repo.UpdateMetric("MSpanSys", m.MSpanSys)
	repo.UpdateMetric("Mallocs", m.Mallocs)
	repo.UpdateMetric("NextGC", m.NextGC)
	repo.UpdateMetric("NumForcedGC", m.NumForcedGC)
	repo.UpdateMetric("NumGC", m.NumGC)
	repo.UpdateMetric("OtherSys", m.OtherSys)
	repo.UpdateMetric("PauseTotalNs", m.PauseTotalNs)
	repo.UpdateMetric("StackInuse", m.StackInuse)
	repo.UpdateMetric("StackSys", m.StackSys)
	repo.UpdateMetric("Sys", m.Sys)
	repo.UpdateMetric("TotalAlloc", m.TotalAlloc)

	repo.UpdateMetric("RandomValue", rand.Float64())
	repo.UpdateMetric("PollCount", int64(1))
	return nil
}

func SendEvent(tick *time.Ticker, repo repository.Repository, httpc http.Client) {
	for range tick.C {
		var err = Send(repo, httpc)
		if err != nil {
			log.Err(err)
		}
	}
}

func Send(repo repository.Repository, httpc http.Client) error {
	var metricType string
	log.Printf("metrics=%s", repo.GetAllAsString())
	for k, v := range repo.GetAll() {
		switch v.(type) {
		case float64:
			metricType = "gauge"
		case int64:
			metricType = "counter"
		default:
			return fmt.Errorf("invalid metric type %s", reflect.TypeOf(v).Name())
		}
		var updateURL = fmt.Sprintf("http://%s/update/%s/%s/%+v", "localhost:8080", metricType, k, v)
		log.Printf("url=%s", updateURL)
		resp, err := httpc.Post(updateURL, "text/plain", nil)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HPPT status code = %d", resp.StatusCode)
		}
		defer resp.Body.Close()
	}
	log.Printf("metrics sended")
	repo.UpdateMetric("PollCount", int64(0))
	return nil
}
