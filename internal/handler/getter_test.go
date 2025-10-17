package handler

import (
	"net/http"
	"testing"
)

var hg MetricGetHandler = NewMetricGetHandler(repo)

func TestGetter(t *testing.T) {
	var zeroStatus = 0
	var w = ResponseWriterMock{status: &zeroStatus}

	TItem(t, hg, "http://localhost:8080/value/", w, http.StatusNotFound, "path too short")
	TItem(t, hg, "http://localhost:8080/value/ololo/ololo/ololo", w, http.StatusBadRequest, "invalid metric type")
	TItem(t, hg, "http://localhost:8080/value/gauge//", w, http.StatusNotFound, "invalid metric name")
	TItem(t, hg, "http://localhost:8080/value/gauge/ololo", w, http.StatusNotFound, "mertic not found")
	TItem(t, hg, "http://localhost:8080/value/gauge/float64", w, http.StatusOK, "float64 value")
	TItem(t, hg, "http://localhost:8080/value/counter/int64", w, http.StatusOK, "int64 value")
}
