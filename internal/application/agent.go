package application

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/metrics"
	"strconv"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"encoding/json"

	models "github.com/dag3322-oss/metrics/internal/model"
	repository "github.com/dag3322-oss/metrics/internal/repository"
	echo "github.com/labstack/echo/v4"
)

type Agent struct {
	Host           string
	ReportInterval *int64
	PollInterval   *int64
}

func (a *Agent) setParams(cmdArgs []string) error {
	var err error
	if cmdArgs == nil {
		cmdArgs = os.Args[1:]
		log.Printf("os.Args=%s", os.Args)
	}
	log.Printf("cmdArgs=%s", cmdArgs)
	var flagSet = flag.NewFlagSet("server", flag.ExitOnError)
	var flagHost = flagSet.String("a", "localhost:8080", "host:port")
	var flagReportInterval = flagSet.Int64("r", 10, "send interval(seconds)")
	var flagPollInterval = flagSet.Int64("p", 2, "poll interval(seconds)")
	if len(cmdArgs) > 0 {
		flagSet.Parse(cmdArgs) //on error will print descriptive error and exit
	}
	log.Printf("before assign s.host=%s,os.host=%s,flag.host=%s", a.Host, os.Getenv("ADDRESS"), *flagHost)

	a.Host = NotEmpty(a.Host, os.Getenv("ADDRESS"), *flagHost)
	log.Printf("after assign s.host=%s", a.Host)
	_, _, err = net.SplitHostPort(a.Host)
	if err != nil {
		return err
	}
	log.Printf("host=%s", a.Host)

	if a.ReportInterval == nil {
		se, exists := os.LookupEnv("REPORT_INTERVAL")
		if exists {
			i, err := strconv.ParseInt(se, 10, 8)
			if err != nil {
				return err
			}
			a.ReportInterval = &i
		} else {
			a.ReportInterval = flagReportInterval
		}
	}
	log.Printf("after assign reportInterval=%d", a.ReportInterval)

	if a.PollInterval == nil {
		se, exists := os.LookupEnv("POLL_INTERVAL")
		if exists {
			i, err := strconv.ParseInt(se, 10, 8)
			if err != nil {
				return err
			}
			a.PollInterval = &i
		} else {
			a.PollInterval = flagPollInterval
		}
	}
	log.Printf("after assign PollInterval=%d", a.PollInterval)

	return nil
}

func (a *Agent) Run(cmdArgs []string) error {
	var err = a.setParams(cmdArgs)
	if err != nil {
		log.Err(err).Msg("setParams exception")
		return err
	}

	var repo = repository.NewMemRepository()

	Collect(time.Now(), repo)
	var collectTimer = time.NewTicker(time.Second * time.Duration(*a.PollInterval))
	go CollectEvent(collectTimer, repo)

	var httpc = http.Client{Timeout: time.Second * time.Duration(30)}
	var sendTimer = time.NewTicker(time.Second * time.Duration(*a.ReportInterval))
	go SendEvent(sendTimer, repo, httpc, a.Host)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	collectTimer.Stop()
	sendTimer.Stop()

	return nil
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

func SendEvent(tick *time.Ticker, repo repository.Repository, httpc http.Client, host string) {
	for range tick.C {
		var err = Send(repo, httpc, host)
		if err != nil {
			log.Err(err).Msg("")
		}
	}
}

func Send(repo repository.Repository, httpc http.Client, host string) error {
	var err error
	var m *models.Metrics
	//log.Printf("metrics=%s", repo.GetAllAsString())
	for k, v := range repo.GetAll() {
		m = new(models.Metrics)
		err = models.FromKeyValue(m, k, v)
		if err != nil {
			log.Err(err).Msg("model create exception")
			return err
		}
		var b []byte
		b, err = json.Marshal(*m)
		if err != nil {
			log.Err(err).Msg("json marshal exception")
			return err
		}

		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		_, err = gzWriter.Write(b)
		if err != nil {
			log.Err(err).Msg("zip exception")
			return err
		}
		gzWriter.Close()

		req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/update", host), &buf)
		if err != nil {
			log.Err(err).Msg("request create exception")
			return err
		}
		req.Header.Set(echo.HeaderContentEncoding, "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Err(err).Msg("request send exception")
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("HTTP status code = %d", resp.StatusCode)
			log.Err(err).Msg("")
			return err
		}
		defer resp.Body.Close()
	}
	log.Printf("metrics sended")
	repo.UpdateMetric("PollCount", int64(0))
	return nil
}
