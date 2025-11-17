package handler

import (
	"net/http"

	metrics "github.com/dag3322-oss/metrics/internal/repository"
	echo "github.com/labstack/echo/v4"
)

type MetricListHandler struct {
	repo metrics.Repository
}

func NewMetricListHandler(
	repo metrics.Repository,
) MetricListHandler {
	return MetricListHandler{repo: repo}
}

func (h MetricListHandler) HandleMetricsList(c echo.Context) error {
	c.Response().Header().Add("Content-Type", "text/html")
	var s = h.repo.GetAllAsString()
	c.Response().Write([]byte(s))
	c.Response().WriteHeader(http.StatusOK)
	return nil
}
