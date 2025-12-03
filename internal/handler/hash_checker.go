package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/dag3322-oss/metrics/internal/helper"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type HashCheckConfig struct {
	Key *string
}

var DefaultHashCheckConfig = HashCheckConfig{}

func HashCheck() echo.MiddlewareFunc {
	return HashCheckWithConfig(DefaultHashCheckConfig)
}

func HashCheckWithConfig(config HashCheckConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Key == nil {
				return next(c)
			}
			headerHash := c.Request().Header.Get(helper.HashHeaderName)
			if headerHash == "" {
				log.Debug().Msg("no sha256 header")
				return next(c)
			}
			r := c.Request().Body
			if r == nil {
				log.Debug().Msg("empty body")
				return next(c)
			}
			defer r.Close()
			b, err := io.ReadAll(r)
			if err != nil {
				log.Err(err).Msg("read body")
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			bodyHash := sha256.Sum256(append(b[:], []byte(*config.Key)...))
			s := base64.StdEncoding.EncodeToString(bodyHash[:])
			if headerHash != s {
				log.Debug().Msg("invalid hash")
				return echo.NewHTTPError(http.StatusBadRequest)
			}
			log.Debug().Msg("hash validated")
			c.Request().Body = io.NopCloser(bytes.NewReader(b))
			return next(c)
		}
	}
}
