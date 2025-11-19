package application

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	handlers "github.com/dag3322-oss/metrics/internal/handler"
	repository "github.com/dag3322-oss/metrics/internal/repository"
	echo "github.com/labstack/echo/v4"
	middleware "github.com/labstack/echo/v4/middleware"
)

type Server struct {
	Host                   string
	StoreInterval          *int64
	StoragePath            string
	LoadFromStorageOnStart *bool
}

func (s *Server) setParams(cmdArgs []string) error {
	var err error
	if cmdArgs == nil {
		cmdArgs = os.Args[1:]
	}
	var flagSet = flag.NewFlagSet("server", flag.ExitOnError)
	var flagHost = flagSet.String("a", "localhost:8080", "host:port")
	var flagStoreInterval = flagSet.Int64("i", 300, "metrics store to file interval, sec")
	var flagStoragePath = flagSet.String("f", "./metrics.json", "metrics storage path")
	var flagLoadFromStorageOnStart = flagSet.Bool("r", false, "load metrics from storage on start")
	if len(cmdArgs) > 0 {
		flagSet.Parse(cmdArgs)
	}
	log.Debug().Msg(fmt.Sprintf("before assign s.host=%s,os.host=%s,flag.host=%s", s.Host, os.Getenv("ADDRESS"), *flagHost))

	s.Host = NotEmpty(s.Host, os.Getenv("ADDRESS"), *flagHost)
	log.Debug().Msg(fmt.Sprintf("after assign s.host=%s", s.Host))
	_, _, err = net.SplitHostPort(s.Host)
	if err != nil {
		return err
	}
	log.Debug().Msg(fmt.Sprintf("host=%s", s.Host))

	if s.StoreInterval == nil {
		se, exists := os.LookupEnv("STORE_INTERVAL")
		if exists {
			i, err := strconv.ParseInt(se, 10, 8)
			if err != nil {
				return err
			}
			s.StoreInterval = &i
		} else {
			s.StoreInterval = flagStoreInterval
		}
	}

	s.StoragePath = NotEmpty(s.StoragePath, os.Getenv("FILE_STORAGE_PATH"), *flagStoragePath)

	if s.LoadFromStorageOnStart == nil {
		lso, exists := os.LookupEnv("RESTORE")
		if exists {
			b2, err := strconv.ParseBool(lso)
			if err != nil {
				return err
			}
			s.LoadFromStorageOnStart = &b2
		} else {
			s.LoadFromStorageOnStart = flagLoadFromStorageOnStart
		}
	}

	return err
}

func (s Server) Run(cmdArgs []string) error {
	var err = s.setParams(cmdArgs)
	if err != nil {
		log.Err(err).Msg("setParams exception")
		return err
	}

	var repo = repository.NewMemRepository()

	e := echo.New()

	e.Use(middleware.Decompress())

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:          true,
		LogMethod:       true,
		LogLatency:      true,
		LogStatus:       true,
		LogResponseSize: true,
		LogHeaders:      []string{"Content-Type", "Content-Length"},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info().
				Str("url", v.URI).
				Str("method", v.Method).
				Int64("duration", v.Latency.Milliseconds()).
				Int("status", v.Status).
				Int64("response_size", v.ResponseSize).
				Strs("content_type", v.Headers["Content-Type"]).
				Strs("content_length", v.Headers["Content-Length"]).
				Strs("content_encoding", v.Headers[echo.HeaderContentEncoding]).
				Str("headers", fmt.Sprintf("%v", v.Headers)).
				Msg("request")
			return nil
		},
	}))

	e.Use(handlers.GzipWithConfig(handlers.GzipConfig{
		Skipper: func(c echo.Context) bool {
			/* mediaType := handlers.GetMediaType(c.Request())
			return !(mediaType == echo.MIMEApplicationJSON || mediaType == echo.MIMETextHTML) */
			return false
		},
		MinLength: 10,
	}))

	hu := handlers.NewMetricUpdateHandler(repo)
	updates := e.Group("/update")
	updates.GET("*", hu.HandleMetricUpdate)
	updates.POST("*", hu.HandleMetricUpdate)

	hl := handlers.NewMetricListHandler(repo)
	lists := e.Group("/")
	lists.GET("*", hl.HandleMetricsList)

	hg := handlers.NewMetricGetHandler(repo)
	values := e.Group("/value")
	values.GET("*", hg.HandleMetricGet)
	values.POST("*", hg.HandleMetricGet)

	log.Debug().Msg(fmt.Sprintf("LoadFromStorageOnStart=%t", *s.LoadFromStorageOnStart))
	if s.LoadFromStorageOnStart != nil && *s.LoadFromStorageOnStart {
		handlers.Load(repo, s.StoragePath)
	}

	if s.StoreInterval != nil {
		var flushTimer = time.NewTicker(time.Second * time.Duration(*s.StoreInterval))
		go FlushEvent(flushTimer, repo, s.StoragePath)
	}

	if err := e.Start(s.Host); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Err(err).Msg("failed to start server")
	}

	return err
}

func FlushEvent(tick *time.Ticker, repo repository.Repository, filePath string) {
	for range tick.C {
		handlers.Flush(repo, filePath)
	}
}
