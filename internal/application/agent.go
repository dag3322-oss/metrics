package application

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"runtime/metrics"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"encoding/json"

	env "github.com/caarlos0/env/v11"

	repository "github.com/dag3322-oss/metrics/internal/repository"
)

type Agent struct {
	host           string
	reportInterval int64
	pollInterval   int64
}

func (a *Agent) SetHost(host string) {
	a.host = host
}

func (a *Agent) SetReportInterval(reportInterval int64) {
	a.reportInterval = reportInterval
}

func (a *Agent) SetPollInterval(pollInterval int64) {
	a.pollInterval = pollInterval
}

type EnvParams struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
}

func (a *Agent) setParams(cmdArgs []string) error {
	var envParams EnvParams
	var err = env.Parse(&envParams)
	if err != nil {
		return err
	}

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
	log.Printf("before assign s.host=%s,os.host=%s,flag.host=%s", a.host, os.Getenv("ADDRESS"), *flagHost)

	if a.host == "" {
		a.host = os.Getenv("ADDRESS")
	}
	if a.host == "" {
		a.host = *flagHost
	}
	log.Printf("after assign host=%s", a.host)
	_, _, err = net.SplitHostPort(a.host)
	log.Printf("host=%s", a.host)
	if err != nil {
		return err
	}

	if a.reportInterval == 0 {
		a.reportInterval = envParams.ReportInterval
	}
	if a.reportInterval == 0 {
		a.reportInterval = *flagReportInterval
	}
	log.Printf("after assign reportInterval=%d", a.reportInterval)

	if a.pollInterval == 0 {
		a.pollInterval = envParams.PollInterval
	}
	if a.pollInterval == 0 {
		a.pollInterval = *flagPollInterval
	}
	log.Printf("after assign pollInterval=%d", a.pollInterval)

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
	var collectTimer = time.NewTicker(time.Second * time.Duration(a.pollInterval))
	go CollectEvent(collectTimer, repo)

	var httpc = http.Client{Timeout: time.Second * time.Duration(30)}
	var sendTimer = time.NewTicker(time.Second * time.Duration(a.reportInterval))
	go SendEvent(sendTimer, repo, httpc, a.host)

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
	//log.Printf("metrics=%s", repo.GetAllAsString())
	var m repository.Metrics
	for k, v := range repo.GetAll() {
		m.ID = k
		switch vt := v.(type) {
		case float64:
			m.MType = "gauge"
			if f, ok := v.(float64); ok {
				m.Value = &f
			} else {
				err = fmt.Errorf("any to float conversion error")
				log.Err(err).Msg("")
				return err
			}
		case int64:
			m.MType = "counter"
			if i, ok := v.(int64); ok {
				m.Delta = &i
			} else {
				err = fmt.Errorf("any to int conversion error")
				log.Err(err).Msg("")
				return err
			}
		default:
			err = fmt.Errorf("invalid metric type %s", reflect.TypeOf(vt).Name())
			log.Err(err).Msg("")
			return err
		}
		var b []byte
		b, err = json.Marshal(m)
		if err != nil {
			log.Err(err).Msg("json marshal exception")
			return err
		}
		resp, err := httpc.Post(
			fmt.Sprintf("http://%s/update", host),
			"application/json",
			bytes.NewBuffer(b),
		)
		if err != nil {
			log.Err(err).Msg("http post exception")
			return err
		}
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
