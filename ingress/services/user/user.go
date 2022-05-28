package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/dssolutioninc/dss_gke/ingress/services/user/handler"
)

// Default Server Port
const DEFAULT_SERVER_PORT = ":8080"

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Set routing
	v1group := e.Group("/users/v1")

	// Route => handler
	v1group.GET("/roles", handler.SimpleWebHandler{}.Roles)

	// Start server
	e.Logger.Fatal(e.Start(DEFAULT_SERVER_PORT))
}
