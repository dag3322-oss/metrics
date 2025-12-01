package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TItem(t *testing.T, h echo.HandlerFunc, _url string, w ResponseWriterMock, status int, message string) {
	request, err := BuildRequest(_url)
	require.True(t, err == nil, "Error build request")
	c := echo.New().NewContext(request, w)
	c.Request().Method = "GET"
	h(c)
	assert.True(t, *w.status == status, message)
}

func BuildRequest(urlString string) (*http.Request, error) {
	var err error
	urlRef, err := url.Parse(urlString)
	strings.Split(urlRef.Path, "/")
	return &http.Request{URL: urlRef}, err
}

func JItem(t *testing.T, h echo.HandlerFunc, _url string, _json string, w ResponseWriterMock, status int, message string) {
	request, err := BuildJSONRequest(_url, _json)
	assert.NoError(t, err, "BuildJsonRequest")
	require.True(t, err == nil, "Error build request")
	c := echo.New().NewContext(request, w)
	c.Request().Method = "POST"
	h(c)
	assert.True(t, *w.status == status, fmt.Sprintf("%s,status=%v", message, *w.status))
}

func BuildJSONRequest(urlString string, _json string) (*http.Request, error) {
	var err error
	urlRef, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	var r = http.Request{}
	r.URL = urlRef
	r.Header = http.Header{}
	r.Header.Set("Content-Type", "application/json")
	stringReader := strings.NewReader(_json)
	r.Body = io.NopCloser(stringReader)
	r.ContentLength = int64(len(_json))
	return &r, nil
}
