package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type SimpleWebHandler struct {
}

func (sh SimpleWebHandler) Index(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!\n")
}

func (sh SimpleWebHandler) Ping(c echo.Context) error {
	return c.String(http.StatusOK, "Pong!\n")
}
