package handler

import (
	"net/http"

	"github.com/dag3322-oss/metrics/internal/repository"
	echo "github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type DBPingHandler struct {
	db *repository.MetricRepositoryDB
}

func NewDBPingHandler(
	db *repository.MetricRepositoryDB,
) DBPingHandler {
	return DBPingHandler{db: db}
}

func (h DBPingHandler) HandlePing(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/html")
	if h.db == nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
	} else {
		err := h.db.Ping()
		if err == nil {
			c.Response().WriteHeader(http.StatusOK)
		} else {
			c.Response().WriteHeader(http.StatusInternalServerError)
			log.Err(err).Msg("database ping error")
		}
	}
	return nil
}
