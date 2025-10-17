package handler

import (
	"net/http"

	metrics "github.com/dag3322-oss/metrics/internal/repository"
)

type MetricListHandler struct {
	repo metrics.Repository
}

func NewMetricListHandler(
	repo metrics.Repository,
) MetricListHandler {
	if repo == nil {
		panic("empty repository")
	}

	return MetricListHandler{repo: repo}
}

func (h MetricListHandler) Handle(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
	res.Header().Add("Content-Type", "text/plain")
	var s = h.repo.GetAllAsString()
	res.Write([]byte(s))
}
