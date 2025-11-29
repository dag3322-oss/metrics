package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/dag3322-oss/metrics/internal/repository"
	"github.com/dag3322-oss/metrics/internal/service"
	echo "github.com/labstack/echo/v4"
)

type MetricUpdateHandler struct {
	repo repository.Metric
}

func NewMetricUpdateHandler(
	repo repository.Metric,
) MetricUpdateHandler {
	return MetricUpdateHandler{repo: repo}
}

func (h MetricUpdateHandler) HandleMetricUpdate(c echo.Context) error {
	var code int
	var body []byte
	var err error

	switch c.Request().Method {
	case "GET":
		code, body, err = h.HandleMetricUpdateURL(c)
	case "POST":
		mediaType := GetMediaType(c.Request())
		switch mediaType {
		case echo.MIMEApplicationJSON:
			code, body, err = h.HandleMetricUpdateJSON(c)
		case echo.MIMETextPlain, "":
			code, body, err = h.HandleMetricUpdateURL(c)
		default:
			c.Response().WriteHeader(http.StatusUnsupportedMediaType)
			return fmt.Errorf("unsupported media type:%s", mediaType)
		}
	default:
		code = http.StatusMethodNotAllowed
		err = fmt.Errorf("unsupported http method:%s", c.Request().Method)
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

func (h MetricUpdateHandler) HandleMetricUpdateURL(c echo.Context) (code int, body []byte, err error) {
	code, name, value, err := ParseURL(*c.Request().URL)
	if code != http.StatusOK {
		return code, nil, err
	}

	m, err := service.NameValueToModel(name, value)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}

	err = h.repo.SetOne(*m)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}

	return code, nil, err
}

func (h MetricUpdateHandler) HandleMetricUpdateJSON(c echo.Context) (code int, body []byte, err error) {
	var m model.Metric
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

	code, err = model.Validate(model.ActionUpdate, &m)
	if code != http.StatusOK {
		return code, nil, err
	}

	err = h.repo.SetOne(m)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}

	return http.StatusOK, nil, nil
}

func (h MetricUpdateHandler) HandleMetricsUpdateJSON(c echo.Context) (code int, body []byte, err error) {
	var m []model.Metric
	var b []byte
	var mm map[string]model.Metric = make(map[string]model.Metric)

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

	for _, item := range m {
		code, err = model.Validate(model.ActionUpdate, &item)
		if code != http.StatusOK {
			return code, nil, err
		}
		mm[item.ID] = item
	}

	err = h.repo.SetList(mm)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}

	return http.StatusOK, nil, nil
}

func (h MetricUpdateHandler) HandleMetricsUpdate(c echo.Context) error {
	var code int
	var body []byte
	var err error

	switch c.Request().Method {
	case "POST":
		mediaType := GetMediaType(c.Request())
		switch mediaType {
		case echo.MIMEApplicationJSON:
			code, body, err = h.HandleMetricsUpdateJSON(c)

		default:
			c.Response().WriteHeader(http.StatusUnsupportedMediaType)
			return fmt.Errorf("unsupported media type:%s", mediaType)
		}
	default:
		code = http.StatusMethodNotAllowed
		err = fmt.Errorf("unsupported http method:%s", c.Request().Method)
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
