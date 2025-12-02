package application

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/rs/zerolog/log"

	handlers "github.com/dag3322-oss/metrics/internal/handler"
	repository "github.com/dag3322-oss/metrics/internal/repository"
	echo "github.com/labstack/echo/v4"
	middleware "github.com/labstack/echo/v4/middleware"
)

var serverReady = make(chan struct{})

type StorageTypeType string

const (
	StorageTypeMem  StorageTypeType = "memory"
	StorageTypeFile StorageTypeType = "file"
	StorageTypeDB   StorageTypeType = "db"
)

type Server struct {
	Host                   string
	StoreInterval          *int64
	StoragePath            *string
	LoadFromStorageOnStart *bool
	DBConnectionString     *string
	StorageType            StorageTypeType
}

func (s *Server) setParams(cmdArgs []string) error {
	var err error

	for _, s := range os.Environ() {
		log.Debug().Msg(fmt.Sprintf("OS_%s", s))
	}

	if cmdArgs == nil {
		cmdArgs = os.Args[1:]
	}
	var flagSet = flag.NewFlagSet("server", flag.ExitOnError)
	var flagHost = flagSet.String("a", "localhost:8080", "host:port")
	var flagStoreInterval = flagSet.Int64("i", 300, "metrics store to file interval, sec")
	var flagStoragePath = flagSet.String("f", "", "metrics storage path")
	var flagLoadFromStorageOnStart = flagSet.Bool("r", false, "load metrics from storage on start")
	var flagDBConnectionString = flagSet.String("d", "", "database connection string")

	if len(cmdArgs) > 0 {
		flagSet.Parse(cmdArgs)
	}
	log.Debug().Str("s.host", s.Host).Str("os.host", os.Getenv("ADDRESS")).Str("flag.host", *flagHost).Msg("before assign")

	s.Host = NotEmpty(s.Host, os.Getenv("ADDRESS"), *flagHost)
	log.Debug().Str("s.host", s.Host).Msg("after assign")
	_, _, err = net.SplitHostPort(s.Host)
	if err != nil {
		return err
	}

	if s.StoreInterval == nil {
		se, exists := os.LookupEnv("STORE_INTERVAL")
		if exists {
			i, err := strconv.ParseInt(se, 10, 8)
			if err != nil {
				return err
			}
			s.StoreInterval = &i
		} else {
			flagSet.Visit(func(f *flag.Flag) {
				if f.Name == "i" {
					s.StoreInterval = flagStoreInterval
					return
				}
			})
		}
	}
	if s.StoreInterval == nil {
		i := int64(300)
		s.StoreInterval = &i
	}

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

	if s.StoragePath == nil {
		fs, exists := os.LookupEnv("FILE_STORAGE_PATH")
		if exists {
			s.StoragePath = &fs
		} else {
			flagSet.Visit(func(f *flag.Flag) {
				if f.Name == "f" {
					s.StoragePath = flagStoragePath
					return
				}
			})
		}
	}

	if s.DBConnectionString == nil {
		db, exists := os.LookupEnv("DATABASE_DSN")
		if exists {
			s.DBConnectionString = &db
		} else {
			flagSet.Visit(func(f *flag.Flag) {
				if f.Name == "d" {
					s.DBConnectionString = flagDBConnectionString
					log.Debug().Str("flagDBConnectionString", *flagDBConnectionString).Msg("")
					return
				}
			})
		}
	}

	if s.DBConnectionString != nil {
		s.StorageType = StorageTypeDB
	} else if s.StoragePath != nil {
		s.StorageType = StorageTypeFile
	} else {
		s.StorageType = StorageTypeMem
	}

	return err
}

func (s Server) Run(cmdArgs []string) error {
	var err = s.setParams(cmdArgs)
	if err != nil {
		log.Err(err).Msg("setParams exception")
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var repo repository.Metric
	db_ctx, db_ctx_f := context.WithCancel(context.Background())
	defer db_ctx_f()
	switch s.StorageType {
	case StorageTypeDB:
		repo, err = repository.NewDBRepository(s.DBConnectionString, db_ctx)
		if err != nil {
			log.Err(err).Msg("database initialization")
			return err
		}
	case StorageTypeMem:
		repo = repository.NewMemRepository()
	case StorageTypeFile:
		repo = repository.NewFileRepository(*s.StoragePath, false)
	default:
		return fmt.Errorf("storage type not set")
	}
	defer repo.Close()
	log.Debug().Any("repository", s.StorageType).Msg("")

	var repoFlush repository.Metric
	var flushStoragePath string
	if s.StorageType != StorageTypeFile {
		if s.StoragePath != nil {
			flushStoragePath = *s.StoragePath
		} else {
			flushStoragePath = "./metrics.json"
		}
		repoFlush = repository.NewFileRepository(flushStoragePath, true)
	}

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
		MinLength: 1,
	}))

	hu := handlers.NewMetricUpdateHandler(repo)
	updates := e.Group("/update")
	updates.GET("*", hu.HandleMetricUpdate)
	updates.POST("*", hu.HandleMetricUpdate)

	batchUpdates := e.Group("/updates")
	batchUpdates.POST("*", hu.HandleMetricsUpdate)

	hl := handlers.NewMetricListHandler(repo)
	lists := e.Group("/")
	lists.GET("*", hl.HandleMetricsList)

	hg := handlers.NewMetricGetHandler(repo)
	values := e.Group("/value")
	values.GET("*", hg.HandleMetricGet)
	values.POST("*", hg.HandleMetricGet)

	var hp handlers.DBPingHandler
	if s.StorageType == StorageTypeDB {
		hp = handlers.NewDBPingHandler(repo.(*repository.MetricRepositoryDB))
	} else {
		hp = handlers.NewDBPingHandler(nil)
	}
	ping := e.Group("/ping")
	ping.GET("*", hp.HandlePing)

	log.Debug().Bool("LoadFromStorageOnStart", *s.LoadFromStorageOnStart).Msg("")
	if s.StorageType != StorageTypeFile && s.LoadFromStorageOnStart != nil && *s.LoadFromStorageOnStart {
		handlers.Load(repo, repoFlush)
	}

	if s.StorageType != StorageTypeFile && s.StoreInterval != nil {
		var flushTimer = time.NewTicker(time.Second * time.Duration(*s.StoreInterval))
		go FlushEvent(flushTimer, repo, repoFlush)
	}

	go func() {
		if err := e.Start(s.Host); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Err(err).Msg("failed to start server")
		}
	}()

	select {
	case <-serverReady:
		fmt.Println("ready")
	case <-ctx.Done():
		log.Info().Msg("shutdown started...")
		ctxt, f := context.WithTimeout(context.Background(), 20*time.Second)
		defer f()
		e.Shutdown(ctxt)
		log.Info().Msg("http server shutdown complete")
		//repo.Close() wait infinite as described in docs
		db_ctx_f()
		time.Sleep(3 * time.Second)
		log.Info().Msg("repository shutdown complete")
		stop()
	}

	return err
}

func FlushEvent(tick *time.Ticker, repo repository.Metric, repoFlush repository.Metric) {
	for range tick.C {
		handlers.Flush(repo, repoFlush)
	}
}
