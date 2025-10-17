package handler

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var repo MockRepository = MockRepository{m: make(map[string]any)}

type MockRepository struct {
	m map[string]any
}

func (s MockRepository) UpdateMetric(name string, value any) error {
	return nil
}

func (s MockRepository) GetAllAsString() string {
	return ""
}

func (s MockRepository) GetAll() map[string]any {
	s.m["int64"] = int64(100500)
	s.m["float64"] = float64(100500.05)
	return s.m
}

func (s MockRepository) Get(name string) (value any, exists bool) {
	switch name {
	case "int64":
		return int64(100500), true
	case "float64":
		return float64(100500.05), true
	default:
		return nil, false
	}
}

type ResponseWriterMock struct {
	status *int
}

func (w ResponseWriterMock) Header() http.Header {
	return make(http.Header)
}

func (w ResponseWriterMock) Write(b []byte) (int, error) {
	return len(b), nil
}

func (w ResponseWriterMock) WriteHeader(statusCode int) {
	*w.status = statusCode
}

func TItem(t *testing.T, h MetricRequestHandler, _url string, w ResponseWriterMock, status int, message string) {
	request, err := BuildRequest(_url)
	require.True(t, err == nil, "Error build request")
	h.Handle(w, request)
	assert.True(t, *w.status == status, message)
}

func BuildRequest(urlString string) (*http.Request, error) {
	var err error
	urlRef, err := url.Parse(urlString)
	return &http.Request{URL: urlRef}, err
}
