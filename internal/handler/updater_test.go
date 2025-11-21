package handler

import (
	"net/http"
	"testing"

	repository "github.com/dag3322-oss/metrics/internal/repository"
)

var hu MetricUpdateHandler = NewMetricUpdateHandler(repository.NewMemRepository())

func TestUpdater(t *testing.T) {
	var zeroStatus = 0
	var w = ResponseWriterMock{status: &zeroStatus}

	TItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update/", w, http.StatusNotFound, "path too short")
	TItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update/ololo/ololo/ololo", w, http.StatusBadRequest, "invalid metric type")
	TItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update/gauge//100", w, http.StatusNotFound, "invalid metric name")
	TItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update/gauge/ololo/ololo", w, http.StatusBadRequest, "invalid float64 value")
	TItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update/counter/ololo/ololo", w, http.StatusBadRequest, "invalid int64 value")
	TItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update/gauge/ololo/100.500", w, http.StatusOK, "float64 saved")
	TItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update/counter/ololo2/100", w, http.StatusOK, "int64 saved")

	JItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update", "{}", w, http.StatusNotFound, "path too short")
	JItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update", "{\"type\": \"ololo\", \"id\": \"ololo\"}", w, http.StatusBadRequest, "invalid metric type")
	JItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update", "{\"type\": \"gauge\", \"id\": \"\", \"value\": 100}", w, http.StatusNotFound, "invalid metric name")
	JItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update", "{\"type\": \"gauge\", \"id\": \"ololo\", \"value\": \"ololo\"}", w, http.StatusBadRequest, "invalid float64 value")
	JItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update", "{\"type\": \"counter\", \"id\": \"ololo\", \"delta\": \"ololo\"}", w, http.StatusBadRequest, "invalid int64 value")
	JItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update", "{\"type\": \"gauge\", \"id\": \"ololo\", \"value\": 100.500}", w, http.StatusOK, "float64 saved")
	JItem(t, hu.HandleMetricUpdate, "http://localhost:8080/update", "{\"type\": \"counter\", \"id\": \"ololo2\", \"delta\": 100}", w, http.StatusOK, "int64 saved")
}
