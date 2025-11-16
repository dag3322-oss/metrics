package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func LoggerMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {

			wrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			ts := time.Now()
			defer func() {
				log.Info().
					Fields(map[string]interface{}{
						"url":           r.URL.Path,
						"method":        r.Method,
						"duration":      time.Now().UnixMilli() - ts.UnixMilli(),
						"status":        wrw.Status(),
						"response_size": wrw.BytesWritten(),
					}).
					Msg("")
			}()
			next.ServeHTTP(wrw, r)
		}
		return http.HandlerFunc(f)
	}
}
