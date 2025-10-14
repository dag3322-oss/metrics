package updater

import (
	"log"
	"net/http"
	"strconv"
	"strings"

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
	var code = http.StatusOK

	log.Printf("url=%s", strings.Trim(req.URL.Path, "/"))
	var elements = strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(elements) == 4 {
		if elements[2] == "" {
			code = http.StatusNotFound
		} else {
			switch elements[1] {
			case "gauge":
				f, err := strconv.ParseFloat(elements[3], 64)
				if err != nil {
					code = http.StatusBadRequest
				} else {
					err = h.repo.UpdateMetric(elements[2], f)
					if err != nil {
						code = http.StatusBadRequest
						log.Printf("err=%s", err.Error())
					}
				}
			case "counter":
				i, err := strconv.ParseInt(elements[3], 0, 64)
				if err != nil {
					code = http.StatusBadRequest
				} else {
					err = h.repo.UpdateMetric(elements[2], i)
					if err != nil {
						code = http.StatusBadRequest
						log.Printf("err=%s", err.Error())
					}
				}
			default:
				code = http.StatusBadRequest
			}
		}
	} else {
		code = http.StatusNotFound
	}
	res.WriteHeader(code)

	log.Printf("code=%d", code)
	for i, s := range elements {
		log.Printf("i=%d,s=%s", i, s)
	}
}
