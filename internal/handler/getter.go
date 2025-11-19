package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	models "github.com/dag3322-oss/metrics/internal/model"
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
	var code int
	var body []byte
	var err error

	switch c.Request().Method {
	case "GET":
		code, body, err = h.HandleMetricGetURL(c)
	case "POST":
		mediaType := GetMediaType(c.Request())
		switch mediaType {
		case echo.MIMEApplicationJSON:
			code, body, err = h.HandleMetricGetJSON(c)
		case echo.MIMETextPlain, "":
			code, body, err = h.HandleMetricGetURL(c)
		default:
			code = http.StatusUnsupportedMediaType
			err = fmt.Errorf("unsupported media type:%s", mediaType)
		}
	default:
		code = http.StatusMethodNotAllowed
		err = fmt.Errorf("unsupported http method:%s", c.Request().Method)
	}

	if code == http.StatusContinue { // json written using echo Context
		return nil
	}
	c.Response().WriteHeader(code)
	if len(body) > 0 {
		c.Response().Write(body)
	}
	if err != nil {
		log.Err(err)
	}

	return err
}

func (h MetricGetHandler) HandleMetricGetURL(c echo.Context) (code int, body []byte, err error) {
	code, name, _, err := ParseURL(*c.Request().URL)
	if code != http.StatusOK {
		return code, nil, err
	}

	value, exists := h.repo.Get(name)
	log.Debug().Msg(fmt.Sprintf("name=%s,value=%v,exists=%v", name, value, exists))
	if exists {
		c.Response().Header().Set("Content-Type", "text/html")
		return http.StatusOK, []byte(fmt.Sprintf("%v", value)), nil
	} else {
		return http.StatusNotFound, nil, nil
	}
}

func (h MetricGetHandler) HandleMetricGetJSON(c echo.Context) (code int, body []byte, err error) {
	var m models.Metrics
	var b []byte

	code = http.StatusBadRequest

	b, err = io.ReadAll(c.Request().Body)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}

	err = json.Unmarshal(b, &m)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}
	log.Debug().Msg(fmt.Sprintf("body=%s", string(b)))

	code, err = models.Validate(models.ActionGet, &m)
	if code != http.StatusOK {
		return code, nil, err
	}

	value, exists := h.repo.Get(m.ID)
	if !exists {
		log.Debug().Msg(fmt.Sprintf("metric not found name=%s", m.ID))
		return http.StatusNotFound, nil, nil
	} else {
		err = models.SetValue(&m, value)
		if err != nil {
			return http.StatusBadRequest, nil, err
		}
	}

	err = c.JSON(http.StatusOK, &m)
	if err == nil {
		return http.StatusContinue, nil, nil
	} else {
		return http.StatusBadRequest, nil, err
	}
}
