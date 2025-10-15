package application

import (
	"net/http"

	"github.com/rs/zerolog/log"

	handlers "github.com/dag3322-oss/metrics/internal/handler"
	metrics "github.com/dag3322-oss/metrics/internal/repository"
	chi "github.com/go-chi/chi/v5"
	chi_mdw "github.com/go-chi/chi/v5/middleware"
)

type Server struct{}

func (s Server) Run() {
	/* 	file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	   	if err != nil {
	   		log.Fatal("Failed to open log file:", err)
	   	}
	   	log.SetOutput(file)

	*/

	var repo = metrics.NewMemRepository()

	var r = chi.NewRouter()
	r.Use(chi_mdw.Logger)

	var hu = handlers.NewMetricUpdateHandler(repo)
	r.HandleFunc(`/update/*`, hu.Handle)

	var hl = handlers.NewMetricListHandler(repo)
	r.HandleFunc(`/`, hl.Handle)

	var hg = handlers.NewMetricGetHandler(repo)
	r.HandleFunc(`/value/*`, hg.Handle)

	var err = http.ListenAndServe(`:8080`, r)
	if err != nil {
		log.Err(err)
	}
}
