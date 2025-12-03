package application

import (
	"bytes"
	"compress/gzip"
	"context"
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
	"time"

	"github.com/rs/zerolog/log"

	"encoding/json"

	"github.com/dag3322-oss/metrics/internal/helper"
	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/dag3322-oss/metrics/internal/repository"
	"github.com/dag3322-oss/metrics/internal/service"
	echo "github.com/labstack/echo/v4"
)

var agentReady = make(chan struct{})

type Agent struct {
	Host           string
	ReportInterval *int64
	PollInterval   *int64
	repo           repository.Metric
	httpClient     *helper.RetryableClient
	HashKey        *string
}

func (a *Agent) setParams(cmdArgs []string) error {
	var err error
	if cmdArgs == nil {
		cmdArgs = os.Args[1:]
	}
	var flagSet = flag.NewFlagSet("server", flag.ExitOnError)
	var flagHost = flagSet.String("a", "localhost:8080", "host:port")
	var flagReportInterval = flagSet.Int64("r", 10, "send interval(seconds)")
	var flagPollInterval = flagSet.Int64("p", 2, "poll interval(seconds)")
	var flagHashKey = flagSet.String("k", "", "hash key")
	if len(cmdArgs) > 0 {
		flagSet.Parse(cmdArgs)
	}

	a.Host = NotEmpty(a.Host, os.Getenv("ADDRESS"), *flagHost)
	_, _, err = net.SplitHostPort(a.Host)
	if err != nil {
		return err
	}
	log.Debug().Str("host", a.Host).Msg("")

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

	if a.HashKey == nil {
		hk, exists := os.LookupEnv("KEY")
		if exists {
			a.HashKey = &hk
		} else {
			flagSet.Visit(func(f *flag.Flag) {
				if f.Name == "k" {
					a.HashKey = flagHashKey
					return
				}
			})
		}
	}

	return nil
}

func (a *Agent) Run(cmdArgs []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var err = a.setParams(cmdArgs)
	if err != nil {
		log.Err(err).Msg("setParams exception")
		return err
	}

	a.repo = repository.NewMemRepository()

	a.httpClient = helper.NewRetryableClient(a.HashKey)

	Collect(time.Now(), a.repo)
	var collectTimer = time.NewTicker(time.Second * time.Duration(*a.PollInterval))
	go CollectEvent(collectTimer, a.repo)

	var sendTimer = time.NewTicker(time.Second * time.Duration(*a.ReportInterval))
	go SendEvent(sendTimer, a.repo, a.httpClient, a.Host)

	select {
	case <-agentReady:
		fmt.Println("ready")
	case <-ctx.Done():
		collectTimer.Stop()
		sendTimer.Stop()
		stop()
	}

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

func CollectEvent(tick *time.Ticker, repo repository.Metric) {
	for t := range tick.C {
		Collect(t, repo)
	}
}

func Collect(t time.Time, repo repository.Metric) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	setMetric(repo, "Alloc", m.Alloc)
	setMetric(repo, "BuckHashSys", m.BuckHashSys)
	setMetric(repo, "Frees", m.Frees)
	setMetric(repo, "GCCPUFraction", m.GCCPUFraction)
	setMetric(repo, "GCSys", m.GCSys)
	setMetric(repo, "HeapAlloc", m.HeapAlloc)
	setMetric(repo, "HeapIdle", m.HeapIdle)
	setMetric(repo, "HeapInuse", m.HeapInuse)
	setMetric(repo, "HeapObjects", m.HeapObjects)
	setMetric(repo, "HeapReleased", m.HeapReleased)
	setMetric(repo, "HeapSys", m.HeapSys)
	setMetric(repo, "LastGC", m.LastGC)
	setMetric(repo, "Lookups", m.Lookups)
	setMetric(repo, "MCacheInuse", m.MCacheInuse)
	setMetric(repo, "MCacheSys", m.MCacheSys)
	setMetric(repo, "MSpanInuse", m.MSpanInuse)
	setMetric(repo, "MSpanSys", m.MSpanSys)
	setMetric(repo, "Mallocs", m.Mallocs)
	setMetric(repo, "NextGC", m.NextGC)
	setMetric(repo, "NumForcedGC", m.NumForcedGC)
	setMetric(repo, "NumGC", m.NumGC)
	setMetric(repo, "OtherSys", m.OtherSys)
	setMetric(repo, "PauseTotalNs", m.PauseTotalNs)
	setMetric(repo, "StackInuse", m.StackInuse)
	setMetric(repo, "StackSys", m.StackSys)
	setMetric(repo, "Sys", m.Sys)
	setMetric(repo, "TotalAlloc", m.TotalAlloc)

	setMetric(repo, "RandomValue", rand.Float64())
	setMetric(repo, "PollCount", int64(1))
	return nil
}

func SendEvent(tick *time.Ticker, repo repository.Metric, httpc *helper.RetryableClient, host string) {
	for range tick.C {
		var err = SendBatch(repo, httpc, host)
		if err != nil {
			log.Err(err).Msg("SendBatch")
		}
	}
}

func Send(repo repository.Metric, httpc *helper.RetryableClient, host string) error {
	var err error
	mm, err := repo.GetAll()
	if err != nil {
		return err
	}
	for _, m := range mm {
		var b []byte
		b, err = json.Marshal(m)
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		_, err = gzWriter.Write(b)
		if err != nil {
			return err
		}
		gzWriter.Close()

		req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/update", host), &buf)
		if err != nil {
			return err
		}
		req.Header.Set(echo.HeaderContentEncoding, "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resp, err := httpc.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP status code = %d", resp.StatusCode)
		}
		defer resp.Body.Close()
	}
	log.Debug().Msg("metrics sended")
	setMetric(repo, "PollCount", int64(0))
	return nil
}

func SendBatch(repo repository.Metric, httpc *helper.RetryableClient, host string) error {
	var err error
	mm, err := repo.GetAll()
	if err != nil {
		return err
	}
	var a []model.Metric
	for _, m := range mm {
		a = append(a, m)
	}

	var b []byte
	b, err = json.Marshal(&a)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err = gzWriter.Write(b)
	if err != nil {
		return err
	}
	gzWriter.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/updates", host), &buf)
	if err != nil {
		return err
	}
	req.Header.Set(echo.HeaderContentEncoding, "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp, err := httpc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status code = %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	log.Debug().Msg("metrics sended")
	err = setMetric(repo, "PollCount", int64(0))
	if err != nil {
		return err
	}

	return nil
}

func setMetric(repo repository.Metric, name string, value any) error {
	m, err := service.NameValueToModel(name, value)
	if err != nil {
		return err
	}
	log.Debug().Fields(m).Msg("setMetric")
	err = repo.SetOne(*m)
	if err != nil {
		return err
	}
	return nil
}
