package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type SimpleWebHandler struct {
}

func (sh SimpleWebHandler) Current(c echo.Context) error {
	return c.String(http.StatusOK, "Current taks")
}
