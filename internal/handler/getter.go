package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/dag3322-oss/metrics/internal/repository"
	echo "github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type MetricGetHandler struct {
	repo repository.Metric
}

func NewMetricGetHandler(
	repo repository.Metric,
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

	m, err := h.repo.Get(name)
	if err != nil {
		return code, nil, err
	}
	if m != nil {
		c.Response().Header().Set("Content-Type", "text/html")
		log.Debug().Msg(fmt.Sprintf("name=%s,value=%s,exists", m.ID, m.StringValue()))
		return http.StatusOK, []byte(fmt.Sprintf("%s", m.StringValue())), nil
	} else {
		log.Debug().Msg(fmt.Sprintf("name=%s,not exists", name))
		return http.StatusNotFound, nil, nil
	}
}

func (h MetricGetHandler) HandleMetricGetJSON(c echo.Context) (code int, body []byte, err error) {
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

	code, err = model.Validate(model.ActionGet, &m)
	if code != http.StatusOK {
		return code, nil, err
	}

	m2, err := h.repo.Get(m.ID)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}
	if m2 == nil {
		log.Debug().Msg(fmt.Sprintf("metric not found name=%s", m.ID))
		return http.StatusNotFound, nil, nil
	}

	err = c.JSON(http.StatusOK, m2)
	if err == nil {
		return http.StatusContinue, nil, nil
	} else {
		return http.StatusBadRequest, nil, err
	}
}
