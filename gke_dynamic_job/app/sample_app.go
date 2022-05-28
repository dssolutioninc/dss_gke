package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/dssolutioninc/dss_gke/k8sclient/app/handler"
)

const (
	DEFAULT_SERVER_PORT = ":8080"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	//index
	e.GET("/index", handler.SampleHandler{}.Index)

	// run a job
	e.POST("/runajob", handler.SampleHandler{}.RunAJob)

	// update job result
	e.POST("/updatejobstatus", handler.SampleHandler{}.UpdateJobStatus)

	//Server Start
	e.Logger.Fatal(e.Start(DEFAULT_SERVER_PORT))
}
