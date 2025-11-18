package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	models "github.com/dag3322-oss/metrics/internal/model"
	metrics "github.com/dag3322-oss/metrics/internal/repository"
	echo "github.com/labstack/echo/v4"
)

type MetricUpdateHandler struct {
	repo metrics.Repository
}

func NewMetricUpdateHandler(
	repo metrics.Repository,
) MetricUpdateHandler {
	if repo == nil {
		panic("empty repository")
	}

	return MetricUpdateHandler{repo: repo}
}

func (h MetricUpdateHandler) HandleMetricUpdate(c echo.Context) error {
	mediaType := GetMediaType(c.Request())
	switch mediaType {
	case echo.MIMEApplicationJSON:
		return h.HandleMetricUpdateJSON(c)
	case echo.MIMETextPlain, "":
		return h.HandleMetricUpdateURL(c)
	default:
		c.Response().WriteHeader(http.StatusUnsupportedMediaType)
		return fmt.Errorf("unsupported media type:%s", mediaType)
	}
}

func (h MetricUpdateHandler) HandleMetricUpdateURL(c echo.Context) error {
	code, name, value, err := ParseURL(*c.Request().URL)

	if code == http.StatusOK {
		err = h.repo.UpdateMetric(name, value)
	}

	if err != nil {
		code = http.StatusBadRequest
		log.Err(err).Msg("repository exception")
	}

	c.Response().WriteHeader(code)
	return err
}

func (h MetricUpdateHandler) HandleMetricUpdateJSON(c echo.Context) error {
	var m models.Metrics
	var code = http.StatusOK
	var name string
	var value any
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
		code, err = models.Validate(models.ActionUpdate, &m)
		if err != nil {
			log.Err(err).Msg("validate exception")
		}
	}
	// по непонятной причине Bind не работает с декомпрессированым телом запроса,
	// причём декомпрессированным как в middleware так и вручную
	// deflate ожидаемо не помог
	/* 	if err = c.Bind(&m); err != nil {
	   		code = http.StatusBadRequest
	   		log.Err(err).Msg("json unmarshall")
	   	}
	*/
	if code == http.StatusOK {
		name, value, err = models.ToKeyValue(&m)
		if err != nil {
			log.Err(err).Msg("validate")
			code = http.StatusBadRequest
		}
	}

	if code == http.StatusOK {
		err = h.repo.UpdateMetric(name, value)
		if err != nil {
			code = http.StatusBadRequest
			log.Err(err).Msg("repository update")
		}
	}

	c.Response().WriteHeader(code)
	return err
}
