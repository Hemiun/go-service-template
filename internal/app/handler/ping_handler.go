package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"go-service-template/internal/app/dto"
	"go-service-template/internal/app/infrastructure"
)

// PingService - service interface for service health check
//
//go:generate mockgen -destination=mocks/mock_ping_service.go -package=mocks . PingService
type PingService interface {
	Ping(ctx context.Context) dto.Ping
	PingWithDelay(ctx context.Context) dto.Ping
}

// PingClient - interface for rest client
//
//go:generate mockgen -destination=mocks/mock_ping_client.go -package=mocks . PingClient
type PingClient interface {
	Ping(ctx context.Context) (*dto.Ping, error)
}

// PingHandler - handler for health checking
type PingHandler struct {
	infrastructure.SugarLogger
	pingService PingService
	pingClient  PingClient
}

// NewPingHandler - return new PingHandler struct
func NewPingHandler(service PingService, client PingClient) *PingHandler {
	var target PingHandler
	target.pingService = service
	target.pingClient = client
	return &target
}

// PingHandler godoc
// @Summary send ping query to service
// @Description Method for service checking
// @Tags
// @Accept json
// @Produce json
// @Success 200  {object} dto.Ping
// Failure 400 {object} httputil.HTTPError
// Failure 500 {object} httputil.HTTPError
// @Router /ping [get]
func (h *PingHandler) PingHandler(c echo.Context) error {
	c.Response().Header().Set("content-type", echo.MIMEApplicationJSON)
	resultDTO := h.pingService.Ping(c.Request().Context())
	err := c.JSON(http.StatusOK, resultDTO)
	return err
}

// PingWithDelayHandler godoc
// @Summary send ping query to service with delay 30 sec
// @Description Method for service checking
// @Tags
// @Accept json
// @Produce json
// @Success 200  {object} dto.Ping
// Failure 400 {object} httputil.HTTPError
// Failure 500 {object} httputil.HTTPError
// @Router /pingwithdelay [get]
func (h *PingHandler) PingWithDelayHandler(c echo.Context) error {
	c.Response().Header().Set("content-type", echo.MIMEApplicationJSON)
	resultDTO := h.pingService.PingWithDelay(c.Request().Context())
	err := c.JSON(http.StatusOK, resultDTO)
	return err
}

// PingViaClient godoc
// @Summary send ping query to yourself via http client
// @Description Method for service checking
// @Tags
// @Accept json
// @Produce json
// @Success 200  {object} dto.Ping
// Failure 400 {object} httputil.HTTPError
// Failure 500 {object} httputil.HTTPError
// @Router /pingviaclient [get]
func (h *PingHandler) PingViaClient(c echo.Context) error {
	currCtx := c.Request().Context()

	c.Response().Header().Set("content-type", echo.MIMEApplicationJSON)
	resultDTO, err := h.pingClient.Ping(currCtx)
	if err != nil {
		h.LogError(currCtx, "can't process request", err)
		return err
	}
	err = c.JSON(http.StatusOK, resultDTO)
	return err
}
