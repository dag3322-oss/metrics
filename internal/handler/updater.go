package handler

import (
	"net/http"

	"github.com/rs/zerolog/log"

	metrics "github.com/dag3322-oss/metrics/internal/repository"
)

type MetricUpdateHandler struct {
	repo metrics.Repository
}

func NewMetricUpdateHandler(
	repo metrics.Repository,
) MetricUpdateHandler {
	if repo == nil {
		panic("empty repository")
	}

	return MetricUpdateHandler{repo: repo}
}

func (h MetricUpdateHandler) Handle(res http.ResponseWriter, req *http.Request) {
	code, name, value, err := ParseURL(*req.URL)

	if code == http.StatusOK {
		err = h.repo.UpdateMetric(name, value)
	}

	if err != nil {
		code = http.StatusBadRequest
		log.Err(err)
	}

	res.WriteHeader(code)
}
