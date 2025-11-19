package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	models "github.com/dag3322-oss/metrics/internal/model"
	"github.com/rs/zerolog/log"
)

func ParseURL(u url.URL) (resultCode int, name string, value any, err error) {
	resultCode = http.StatusOK

	log.Debug().Msg(fmt.Sprintf("url path=%s", strings.Trim(u.Path, "/")))
	var elements = strings.Split(strings.Trim(u.Path, "/"), "/")
	var action = elements[0]
	if (action == "update" && len(elements) >= 4) || (action == "value" && len(elements) >= 3) {
		name = elements[2]
		if name == "" {
			resultCode = http.StatusNotFound
		} else {
			switch elements[1] {
			case models.Gauge:
				if action == "update" {
					f, err := strconv.ParseFloat(elements[3], 64)
					if err != nil {
						resultCode = http.StatusBadRequest
					} else {
						value = f
					}
				}
			case models.Counter:
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
