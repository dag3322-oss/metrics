package handler

import (
	"net/http"
	"strings"

	echo "github.com/labstack/echo/v4"
)

func GetMediaType(req *http.Request) string {
	base, _, _ := strings.Cut(req.Header.Get(echo.HeaderContentType), ";")
	return strings.TrimSpace(base)

}
