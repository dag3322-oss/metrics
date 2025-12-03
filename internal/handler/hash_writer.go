package handler

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"net"
	"net/http"

	"github.com/dag3322-oss/metrics/internal/helper"
	"github.com/labstack/echo/v4"
)

type HashWriteConfig struct {
	Key *string
}

var DefaultHashWriteConfig = HashWriteConfig{}

func HashWrite() echo.MiddlewareFunc {
	return HashWriteWithConfig(DefaultHashWriteConfig)
}

type hashResponseWriter struct {
	w   http.ResponseWriter
	key *string
}

func HashWriteWithConfig(config HashWriteConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rw := &hashResponseWriter{w: c.Response().Writer, key: config.Key}
			c.Response().Writer = rw
			return next(c)
		}
	}
}

func (w *hashResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *hashResponseWriter) WriteHeader(code int) {
	w.w.WriteHeader(code)
}

func (w *hashResponseWriter) Write(b []byte) (int, error) {
	bodyHash := sha256.Sum256(append(b[:], []byte(*w.key)...))
	s := base64.StdEncoding.EncodeToString(bodyHash[:])
	w.Header().Set(helper.HashHeaderName, s)
	return w.w.Write(b)
}

func (w *hashResponseWriter) Unwrap() http.ResponseWriter {
	return w.w
}

func (w *hashResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(w.w).Hijack()
}

func (w *hashResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.w.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}
