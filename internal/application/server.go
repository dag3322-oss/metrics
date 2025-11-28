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
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"

	handlers "github.com/dag3322-oss/metrics/internal/handler"
	repository "github.com/dag3322-oss/metrics/internal/repository"
	"github.com/dag3322-oss/metrics/migrations"
	pg_pool "github.com/jackc/pgx/v5/pgxpool"
	echo "github.com/labstack/echo/v4"
	middleware "github.com/labstack/echo/v4/middleware"
)

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
					log.Debug().Msg(fmt.Sprintf("flagDBConnectionString=%s", *flagDBConnectionString))
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

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	var db *pg_pool.Pool
	if s.DBConnectionString != nil {
		db, err = pg_pool.New(context.Background(), *s.DBConnectionString)
		if err != nil {
			log.Err(err).Msg("database connection error")
			return err
		}
		defer db.Close()
	}

	var repo repository.Metric
	switch s.StorageType {
	case StorageTypeDB:
		repo = repository.NewDBRepository(db, context.Background())

		log.Debug().Msg("Database migrations will be applied")
		driver, err := iofs.New(migrations.FS, "sql")
		if err != nil {
			log.Err(err).Msg("iofs driver creation")
			return err
		}
		log.Debug().Msg("iofs driver created")

		m, err := migrate.NewWithSourceInstance("iofs", driver, *s.DBConnectionString)
		if err != nil {
			log.Err(err).Msg("migration instance creation")
			return err
		}
		log.Debug().Msg("migration isnstance created")

		log.Err(err)
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Err(err)
			return err
		}
		log.Debug().Msg("Database migrations applied succesfully")
	case StorageTypeMem:
		repo = repository.NewMemRepository()
	case StorageTypeFile:
		repo = repository.NewFileRepository(*s.StoragePath, false)
	default:
		return fmt.Errorf("storage type not set")
	}
	log.Debug().Msg(fmt.Sprintf("repository=%s", s.StorageType))

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

	hl := handlers.NewMetricListHandler(repo)
	lists := e.Group("/")
	lists.GET("*", hl.HandleMetricsList)

	hg := handlers.NewMetricGetHandler(repo)
	values := e.Group("/value")
	values.GET("*", hg.HandleMetricGet)
	values.POST("*", hg.HandleMetricGet)

	hp := handlers.NewDBPingHandler(db)
	ping := e.Group("/ping")
	ping.GET("*", hp.HandlePing)

	log.Debug().Msg(fmt.Sprintf("LoadFromStorageOnStart=%t", *s.LoadFromStorageOnStart))
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
	<-stopChan
	log.Info().Msg("shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = e.Shutdown(ctx)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
	return err
}

func FlushEvent(tick *time.Ticker, repo repository.Metric, repoFlush repository.Metric) {
	for range tick.C {
		handlers.Flush(repo, repoFlush)
	}
}
