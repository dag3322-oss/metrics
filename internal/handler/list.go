package handler

import (
	"fmt"
	"net/http"

	metrics "github.com/dag3322-oss/metrics/internal/repository"
	echo "github.com/labstack/echo/v4"
)

type MetricListHandler struct {
	repo metrics.Metric
}

func NewMetricListHandler(
	repo metrics.Metric,
) MetricListHandler {
	return MetricListHandler{repo: repo}
}

func (h MetricListHandler) HandleMetricsList(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/html")

	m, err := h.repo.GetAll()
	if err != nil {
		return err
	}
	s := ""
	for _, mm := range m {
		if s != "" {
			s = s + "\n"
		}
		s = s + fmt.Sprintf("%s %s", mm.ID, mm.StringValue())
	}
	c.Response().Write([]byte(s))
	c.Response().WriteHeader(http.StatusOK)
	return nil
}
