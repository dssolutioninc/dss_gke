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

func (sh SimpleWebHandler) Public(c echo.Context) error {
	return c.String(http.StatusOK, "Public!\n")
}

func (sh SimpleWebHandler) Private(c echo.Context) error {
	return c.String(http.StatusOK, "Private!\n")
}
