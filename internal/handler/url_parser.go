package handler

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	metrics "github.com/dag3322-oss/metrics/internal/repository"
	"github.com/rs/zerolog/log"
)

func ParseURL(u url.URL) (resultCode int, name string, value any, err error) {
	resultCode = http.StatusOK

	log.Printf("url path=%s", strings.Trim(u.Path, "/"))
	var elements = strings.Split(strings.Trim(u.Path, "/"), "/")
	var action = elements[0]
	if (action == "update" && len(elements) >= 4) || (action == "value" && len(elements) >= 3) {
		name = elements[2]
		if name == "" {
			resultCode = http.StatusNotFound
		} else {
			switch elements[1] {
			case "gauge":
				if action == "update" {
					f, err := strconv.ParseFloat(elements[3], 64)
					if err != nil {
						resultCode = http.StatusBadRequest
					} else {
						value = f
					}
				}
			case "counter":
				if action == "update" {
					i, err := strconv.ParseInt(elements[3], 0, 64)
					if err != nil {
						resultCode = http.StatusBadRequest
					} else {
						value = i
					}
				}
			default:
				resultCode = http.StatusBadRequest
			}
		}
	} else {
		resultCode = http.StatusNotFound
	}
	return resultCode, name, value, err
}

func Validate(action string, m metrics.Metrics) (resultCode int, name string, value any, err error) {
	log.Printf("model=%+v", m)
	resultCode = http.StatusOK
	name = m.ID
	if name == "" {
		resultCode = http.StatusNotFound
	} else {
		switch m.MType {
		case "gauge":
			if action == "update" {
				if m.Value == nil {
					resultCode = http.StatusBadRequest
				} else {
					value = *m.Value
				}
			}
		case "counter":
			if action == "update" {
				if m.Delta == nil {
					resultCode = http.StatusBadRequest
				} else {
					value = *m.Delta
				}
			}
		default:
			resultCode = http.StatusBadRequest
		}
	}
	return resultCode, name, value, err
}
