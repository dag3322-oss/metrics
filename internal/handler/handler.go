package handler

import "net/http"

type MetricRequestHandler interface {
	Handle(res http.ResponseWriter, req *http.Request)
}
