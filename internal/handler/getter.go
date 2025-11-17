package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	metrics "github.com/dag3322-oss/metrics/internal/repository"
	echo "github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type MetricGetHandler struct {
	repo metrics.Repository
}

func NewMetricGetHandler(
	repo metrics.Repository,
) MetricGetHandler {
	return MetricGetHandler{repo: repo}
}

func (h MetricGetHandler) HandleMetricGet(c echo.Context) error {
	mediaType := GetMediaType(c.Request())
	switch mediaType {
	case echo.MIMEApplicationJSON:
		return h.HandleMetricGetJSON(c)
	case echo.MIMETextPlain, "":
		return h.HandleMetricGetURL(c)
	default:
		c.Response().WriteHeader(http.StatusUnsupportedMediaType)
		return fmt.Errorf("unsupported media type:%s", mediaType)
	}
}

func (h MetricGetHandler) HandleMetricGetURL(c echo.Context) error {
	code, name, _, err := ParseURL(*c.Request().URL)

	if err != nil {
		code = http.StatusBadRequest
		log.Err(err).Msg("parse exception")
	}

	if code == http.StatusOK {
		var value, exists = h.repo.Get(name)
		log.Printf("name=%s,value=%v,exists=%v", name, value, exists)
		if exists {
			c.Response().Header().Add("Content-Type", "text/plain")
			c.Response().Write([]byte(fmt.Sprintf("%v", value)))
		} else {
			code = http.StatusNotFound
		}
	}

	c.Response().WriteHeader(code)
	return err
}

func (h MetricGetHandler) HandleMetricGetJSON(c echo.Context) error {
	var m metrics.Metrics
	var code = http.StatusOK
	var name string
	var err error
	var b []byte

	if code == http.StatusOK {
		b, err = io.ReadAll(c.Request().Body)
		if err != nil {
			code = http.StatusBadRequest
			log.Err(err).Msg("read request body")
		}
	}

	if code == http.StatusOK {
		err = json.Unmarshal(b, &m)
		if err != nil {
			code = http.StatusBadRequest
			log.Err(err).Msg("json unmarshal exception")
		}
		log.Printf("body=%s", string(b))
	}

	if code == http.StatusOK {
		code, name, _, err = Validate("value", m)
		if err != nil {
			code = http.StatusBadRequest
			log.Err(err).Msg("validate exception")
		}
	}

	if code == http.StatusOK {
		var value, exists = h.repo.Get(name)
		log.Printf("name=%s,value=%v,exists=%v", name, value, exists)
		if exists {
			switch m.MType {
			case "gauge":
				if f, ok := value.(float64); ok {
					m.Value = &f
				} else {
					code = http.StatusBadRequest
					log.Err(err).Msg("any to float conversion error")
				}
			case "counter":
				if i, ok := value.(int64); ok {
					m.Delta = &i
				} else {
					code = http.StatusBadRequest
					log.Err(err).Msg("any to int conversion error")
				}
			default:
				code = http.StatusBadRequest
				log.Err(err).Msg("invalid value type")
			}
		} else {
			code = http.StatusNotFound
		}
	}

	if code == http.StatusOK {
		err = c.JSON(code, &m)
		if err != nil {
			code = http.StatusBadRequest
			log.Err(err).Msg("response to json exception")
		}
	}
	c.Response().WriteHeader(code)
	return err
}
