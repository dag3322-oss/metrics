package handler

import (
	"net/http"
	"testing"

	"github.com/dag3322-oss/metrics/internal/repository"
)

var hg MetricGetHandler = NewMetricGetHandler(repository.NewMockRepository())

func TestGetter(t *testing.T) {
	var zeroStatus = 0
	var w = ResponseWriterMock{status: &zeroStatus}

	TItem(t, hg.HandleMetricGet, "http://localhost:8080/value/", w, http.StatusNotFound, "path too short")
	TItem(t, hg.HandleMetricGet, "http://localhost:8080/value/ololo/ololo/ololo", w, http.StatusBadRequest, "invalid metric type")
	TItem(t, hg.HandleMetricGet, "http://localhost:8080/value/gauge//", w, http.StatusNotFound, "invalid metric name")
	TItem(t, hg.HandleMetricGet, "http://localhost:8080/value/gauge/ololo", w, http.StatusNotFound, "metric not found")
	TItem(t, hg.HandleMetricGet, "http://localhost:8080/value/gauge/float64", w, http.StatusOK, "float64 value")
	TItem(t, hg.HandleMetricGet, "http://localhost:8080/value/counter/int64", w, http.StatusOK, "int64 value")

	JItem(t, hg.HandleMetricGet, "http://localhost:8080/value", "{}", w, http.StatusNotFound, "path too short")
	JItem(t, hg.HandleMetricGet, "http://localhost:8080/value", "{\"type\": \"ololo\", \"id\": \"ololo\"}", w, http.StatusBadRequest, "invalid metric type")
	JItem(t, hg.HandleMetricGet, "http://localhost:8080/value", "{\"type\": \"gauge\", \"id\": \"\"}", w, http.StatusNotFound, "invalid metric name")
	JItem(t, hg.HandleMetricGet, "http://localhost:8080/value", "{\"type\": \"gauge\", \"id\": \"ololo\"}", w, http.StatusNotFound, "metric not found")
	JItem(t, hg.HandleMetricGet, "http://localhost:8080/value", "{\"type\": \"gauge\", \"id\": \"float64\"}", w, http.StatusOK, "float64 value")
	JItem(t, hg.HandleMetricGet, "http://localhost:8080/value", "{\"type\": \"counter\", \"id\": \"int64\"}", w, http.StatusOK, "int64 value")
}
