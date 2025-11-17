package application

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"

	handlers "github.com/dag3322-oss/metrics/internal/handler"
	metrics "github.com/dag3322-oss/metrics/internal/repository"
	echo "github.com/labstack/echo/v4"
	middleware "github.com/labstack/echo/v4/middleware"
)

type Server struct {
	host string
}

func (s *Server) SetHost(host string) {
	s.host = host
}

func (s *Server) setParams(cmdArgs []string) error {
	if cmdArgs == nil {
		cmdArgs = os.Args[1:]
		log.Printf("os.Args=%s", os.Args)
	}
	log.Printf("cmdArgs=%s", cmdArgs)
	var flagSet = flag.NewFlagSet("server", flag.ExitOnError)
	var flagHost = flagSet.String("a", "localhost:8080", "host:port")
	if len(cmdArgs) > 0 {
		flagSet.Parse(cmdArgs) //on error will print descriptive error and exit
	}
	log.Printf("before assign s.host=%s,os.host=%s,flag.host=%s", s.host, os.Getenv("ADDRESS"), *flagHost)

	if s.host == "" {
		s.host = os.Getenv("ADDRESS")
	}
	if s.host == "" {
		s.host = *flagHost
	}
	log.Printf("after assign s.host=%s", s.host)
	_, _, err := net.SplitHostPort(s.host)
	log.Printf("host=%s", s.host)
	return err
}

func (s Server) Run(cmdArgs []string) error {
	var err = s.setParams(cmdArgs)
	if err != nil {
		log.Err(err).Msg("setParams exception")
		return err
	}

	var repo = metrics.NewMemRepository()

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

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			mediaType := handlers.GetMediaType(c.Request())
			return !(mediaType == echo.MIMEApplicationJSON || mediaType == echo.MIMETextHTML)
		},
		MinLength: 1,
	}))

	hu := handlers.NewMetricUpdateHandler(repo)
	updates := e.Group("/update")
	updates.GET("*", hu.HandleMetricUpdateURL)
	updates.POST("*", hu.HandleMetricUpdate)

	hl := handlers.NewMetricListHandler(repo)
	lists := e.Group("/")
	lists.GET("*", hl.HandleMetricsList)

	hg := handlers.NewMetricGetHandler(repo)
	values := e.Group("/value")
	values.GET("*", hg.HandleMetricGetURL)
	values.POST("*", hg.HandleMetricGet)

	if err := e.Start(s.host); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Err(err).Msg("failed to start server")
	}
	return err
}
