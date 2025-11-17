package handler

import (
	"net/http"
	"testing"
)

var hu MetricUpdateHandler = NewMetricUpdateHandler(repo)

func TestUpdater(t *testing.T) {
	var zeroStatus = 0
	var w = ResponseWriterMock{status: &zeroStatus}

	TItem(t, hu, "http://localhost:8080/update/", w, http.StatusNotFound, "path too short")
	TItem(t, hu, "http://localhost:8080/update/ololo/ololo/ololo", w, http.StatusBadRequest, "invalid metric type")
	TItem(t, hu, "http://localhost:8080/update/gauge//100", w, http.StatusNotFound, "invalid metric name")
	TItem(t, hu, "http://localhost:8080/update/gauge/ololo/ololo", w, http.StatusBadRequest, "invalid float64 value")
	TItem(t, hu, "http://localhost:8080/update/counter/ololo/ololo", w, http.StatusBadRequest, "invalid int64 value")
	TItem(t, hu, "http://localhost:8080/update/gauge/ololo/100.500", w, http.StatusOK, "float64 saved")
	TItem(t, hu, "http://localhost:8080/update/counter/ololo/100", w, http.StatusOK, "int64 saved")
}
