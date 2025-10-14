package updater

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
}

func (s MockRepository) UpdateMetric(name string, value any) error {
	return nil
}

func (s MockRepository) GetAllAsString() string {
	return ""
}

func (s MockRepository) GetAll() map[string]any {
	return make(map[string]any)
}

type ResponseWriterMock struct {
	status *int
}

func (w ResponseWriterMock) Header() http.Header {
	return nil
}

func (w ResponseWriterMock) Write([]byte) (int, error) {
	return 0, nil
}

func (w ResponseWriterMock) WriteHeader(statusCode int) {
	*w.status = statusCode
}

var repo MockRepository = MockRepository{}

var hu MetricUpdateHandler = NewMetricUpdateHandler(repo)

func TestUpdater(t *testing.T) {
	var zeroStatus = 0
	var w = ResponseWriterMock{status: &zeroStatus}

	TItem(t, "http://localhost:8080/update/", w, http.StatusNotFound, "path too short")
	TItem(t, "http://localhost:8080/update/ololo/ololo/ololo", w, http.StatusBadRequest, "invalid metric type")
	TItem(t, "http://localhost:8080/update/gauge//100", w, http.StatusNotFound, "invalid metric name")
	TItem(t, "http://localhost:8080/update/gauge/ololo/ololo", w, http.StatusBadRequest, "invalid float64 value")
	TItem(t, "http://localhost:8080/update/counter/ololo/ololo", w, http.StatusBadRequest, "invalid int64 value")
	TItem(t, "http://localhost:8080/update/gauge/ololo/100.500", w, http.StatusOK, "float64 saved")
	TItem(t, "http://localhost:8080/update/counter/ololo/100", w, http.StatusOK, "int64 saved")
}

func TItem(t *testing.T, _url string, w ResponseWriterMock, status int, message string) {
	request, err := BuildRequest(_url)
	require.True(t, err == nil, "Error build request")
	hu.Handle(w, request)
	assert.True(t, *w.status == status, message)
}

func BuildRequest(urlString string) (*http.Request, error) {
	var err error
	urlRef, err := url.Parse(urlString)
	return &http.Request{URL: urlRef}, err
}
