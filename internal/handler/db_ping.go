package handler

import (
	"net/http"

	pg_pool "github.com/jackc/pgx/v5/pgxpool"
	echo "github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type DBPingHandler struct {
	db *pg_pool.Pool
}

func NewDBPingHandler(
	db *pg_pool.Pool,
) DBPingHandler {
	return DBPingHandler{db: db}
}

func (h DBPingHandler) HandlePing(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/html")
	if h.db == nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
	} else {
		err := h.db.Ping(c.Request().Context())
		if err == nil {
			c.Response().WriteHeader(http.StatusOK)
		} else {
			c.Response().WriteHeader(http.StatusInternalServerError)
			log.Err(err).Msg("database ping error")
		}
	}
	return nil
}
