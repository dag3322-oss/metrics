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

	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/dag3322-oss/metrics/internal/repository"
	"github.com/dag3322-oss/metrics/internal/service"
	echo "github.com/labstack/echo/v4"
)

type Agent struct {
	Host           string
	ReportInterval *int64
	PollInterval   *int64
	repo           repository.Metric
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
	if len(cmdArgs) > 0 {
		flagSet.Parse(cmdArgs)
	}

	a.Host = NotEmpty(a.Host, os.Getenv("ADDRESS"), *flagHost)
	_, _, err = net.SplitHostPort(a.Host)
	if err != nil {
		return err
	}
	log.Debug().Msg(fmt.Sprintf("host=%s", a.Host))

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

	return nil
}

func (a *Agent) Run(cmdArgs []string) error {
	var err = a.setParams(cmdArgs)
	if err != nil {
		log.Err(err).Msg("setParams exception")
		return err
	}

	a.repo = repository.NewMemRepository()

	Collect(time.Now(), a.repo)
	var collectTimer = time.NewTicker(time.Second * time.Duration(*a.PollInterval))
	go CollectEvent(collectTimer, a.repo)

	var httpc = http.Client{Timeout: time.Second * time.Duration(30)}
	var sendTimer = time.NewTicker(time.Second * time.Duration(*a.ReportInterval))
	go SendEvent(sendTimer, a.repo, httpc, a.Host)

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

func SendEvent(tick *time.Ticker, repo repository.Metric, httpc http.Client, host string) {
	for range tick.C {
		var err = SendBatch(repo, httpc, host)
		if err != nil {
			log.Err(err).Msg("SendBatch")
		}
	}
}

func httpDo(req *http.Request) (resp *http.Response, err error) {
	client := &http.Client{}
	client.Timeout = 30 * time.Second
	i := 0
	for {
		time.Sleep(time.Duration(i) * time.Second)
		resp, err = client.Do(req)
		if err != nil {
			if _, ok := err.(net.Error); ok {
				switch i {
				case 0:
					i = 1
				default:
					i = i + 2
				}
				if i <= 5 {
					log.Debug().Msg(fmt.Sprintf("repeat after timeout delay=%d", i))
					continue
				}
			} else {
				log.Debug().Msg("not network error")
			}
			log.Err(err).Msg("request send exception")
			return nil, err
		}
		return resp, nil
	}

}

func Send(repo repository.Metric, httpc http.Client, host string) error {
	var err error
	mm, err := repo.GetAll()
	if err != nil {
		log.Err(err).Msg("mem repository exception")
		return err
	}
	for _, m := range mm {
		var b []byte
		b, err = json.Marshal(m)
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
		resp, err := httpDo(req)
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
	log.Debug().Msg("metrics sended")
	setMetric(repo, "PollCount", int64(0))
	return nil
}

func SendBatch(repo repository.Metric, httpc http.Client, host string) error {
	var err error
	mm, err := repo.GetAll()
	if err != nil {
		log.Err(err).Msg("mem repository exception")
		return err
	}
	var a []model.Metric
	for _, m := range mm {
		a = append(a, m)
	}

	var b []byte
	b, err = json.Marshal(&a)
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

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/updates", host), &buf)
	if err != nil {
		log.Err(err).Msg("request create exception")
		return err
	}
	req.Header.Set(echo.HeaderContentEncoding, "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp, err := httpDo(req)
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

	log.Debug().Msg("metrics sended")
	setMetric(repo, "PollCount", int64(0))
	return nil
}

func setMetric(repo repository.Metric, name string, value any) {
	m, err := service.NameValueToModel(name, value)
	if err != nil {
		log.Err(err).Msg("NameValueToModel")
		return
	}
	log.Debug().Msg(fmt.Sprintf("setMetric model=%+v", *m))
	repo.SetOne(*m)
}
