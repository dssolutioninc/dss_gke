package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/itdevsamurai/gke/sercure_k8s_service_sa/webapp/handler"
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

	// Route => handler
	e.GET("/index", handler.SimpleWebHandler{}.Index)
	e.GET("/public", handler.SimpleWebHandler{}.Public)
	e.GET("/private", handler.SimpleWebHandler{}.Private)

	// Start server
	e.Logger.Fatal(e.Start(DEFAULT_SERVER_PORT))
}
