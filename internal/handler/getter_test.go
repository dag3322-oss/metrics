package handler

import (
	"net/http"
	"testing"

	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/dag3322-oss/metrics/internal/repository"
)

// mockStringer.EXPECT().String().Return("mockery")

func TestGetter(t *testing.T) {
	var zeroStatus = 0
	var w = ResponseWriterMock{status: &zeroStatus}

	var repo = repository.NewMockMetric(t)

	var hg = NewMetricGetHandler(repo)

	TItem(t, hg.HandleMetricGet, "http://localhost:8080/value/", w, http.StatusNotFound, "path too short")
	TItem(t, hg.HandleMetricGet, "http://localhost:8080/value/ololo/ololo/ololo", w, http.StatusBadRequest, "invalid metric type")
	TItem(t, hg.HandleMetricGet, "http://localhost:8080/value/gauge//", w, http.StatusNotFound, "invalid metric name")
	repo.On("Get", "ololo").Return(nil, nil)
	TItem(t, hg.HandleMetricGet, "http://localhost:8080/value/gauge/ololo", w, http.StatusNotFound, "metric not found")
	f := float64(100500.05)
	var m = model.Metric{ID: "float64", MType: model.Gauge, Value: &f}
	repo.On("Get", "float64").Return(&m, nil)
	TItem(t, hg.HandleMetricGet, "http://localhost:8080/value/gauge/float64", w, http.StatusOK, "float64 value")
	i := int64(100500)
	m = model.Metric{ID: "int64", MType: model.Counter, Delta: &i}
	repo.On("Get", "int64").Return(&m, nil)
	TItem(t, hg.HandleMetricGet, "http://localhost:8080/value/counter/int64", w, http.StatusOK, "int64 value")

	JItem(t, hg.HandleMetricGet, "http://localhost:8080/value", "{}", w, http.StatusNotFound, "path too short")
	JItem(t, hg.HandleMetricGet, "http://localhost:8080/value", "{\"type\": \"ololo\", \"id\": \"ololo\"}", w, http.StatusBadRequest, "invalid metric type")
	JItem(t, hg.HandleMetricGet, "http://localhost:8080/value", "{\"type\": \"gauge\", \"id\": \"\"}", w, http.StatusNotFound, "invalid metric name")
	JItem(t, hg.HandleMetricGet, "http://localhost:8080/value", "{\"type\": \"gauge\", \"id\": \"ololo\"}", w, http.StatusNotFound, "metric not found")
	JItem(t, hg.HandleMetricGet, "http://localhost:8080/value", "{\"type\": \"gauge\", \"id\": \"float64\"}", w, http.StatusOK, "float64 value")
	JItem(t, hg.HandleMetricGet, "http://localhost:8080/value", "{\"type\": \"counter\", \"id\": \"int64\"}", w, http.StatusOK, "int64 value")
}
