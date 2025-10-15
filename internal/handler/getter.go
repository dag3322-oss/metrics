package handler

import (
	"fmt"
	"net/http"

	metrics "github.com/dag3322-oss/metrics/internal/repository"
	"github.com/rs/zerolog/log"
)

type MetricGetHandler struct {
	repo metrics.Repository
}

func NewMetricGetHandler(
	repo metrics.Repository,
) MetricGetHandler {
	if repo == nil {
		panic("empty repository")
	}

	return MetricGetHandler{repo: repo}
}

func (h MetricGetHandler) Handle(res http.ResponseWriter, req *http.Request) {
	code, name, _, err := ParseURL(*req.URL)

	if err != nil {
		code = http.StatusBadRequest
		log.Err(err)
		return
	}

	if code == http.StatusOK {
		var value, exists = h.repo.Get(name)
		log.Printf("name=%s,value=%v,exists=%v", name, value, exists)
		if exists {
			res.Header().Add("Content-Type", "text/plain")
			res.Write([]byte(fmt.Sprintf("%v", value)))
		} else {
			code = http.StatusNotFound
		}
	}

	res.WriteHeader(code)
}
