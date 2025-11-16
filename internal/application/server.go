package application

import (
	"flag"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/dag3322-oss/metrics/internal/handler"
	handlers "github.com/dag3322-oss/metrics/internal/handler"
	metrics "github.com/dag3322-oss/metrics/internal/repository"
	chi "github.com/go-chi/chi/v5"
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

	var r = chi.NewRouter()
	r.Use(handler.LoggerMiddleware())

	var hu = handlers.NewMetricUpdateHandler(repo)
	r.HandleFunc(`/update/*`, hu.Handle)

	var hl = handlers.NewMetricListHandler(repo)
	r.HandleFunc(`/`, hl.Handle)

	var hg = handlers.NewMetricGetHandler(repo)
	r.HandleFunc(`/value/*`, hg.Handle)

	err = http.ListenAndServe(s.host, r)
	if err != nil {
		log.Err(err).Msg("ListenAndServe exception")
	}
	return err
}
