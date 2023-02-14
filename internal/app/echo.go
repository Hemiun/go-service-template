package app

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go-service-template/internal/app/handler"
	echoMiddleware "go-service-template/internal/app/infrastructure/echo"
)

func prepareEcho() {
	e = echo.New()
	e.Pre(echoMiddleware.PrepareRequestID)
	e.Pre(echoMiddleware.PrepareCorrelationID)
	e.Pre(echoMiddleware.PrepareLogger)
	e.Use(echoMiddleware.RequestLogger)
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.HTTPErrorHandler = handler.ErrorHandler
}

func prepareRoutes() {
	v1 := e.Group("/api/v1")
	v1.GET("/ping", pingHandler.PingHandler)
	v1.GET("/pingwithdelay", pingHandler.PingWithDelayHandler)
	v1.GET("/pingviaclient", pingHandler.PingViaClient)
	e.GET("/swagger-ui/*", echoSwagger.WrapHandler)
}
