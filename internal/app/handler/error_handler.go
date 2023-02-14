package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go-service-template/internal/app/dto"
	"go-service-template/internal/app/infrastructure"
)

// ErrorHandler creates new error handler. It handles errors in custom way
func ErrorHandler(incomingError error, c echo.Context) {
	if c.Response().Committed || incomingError == nil {
		return
	}

	log := infrastructure.GetBaseLogger(c.Request().Context())
	var (
		errorDTO *dto.ErrorDTO
		status   int
		cause    error
		err      error
	)

	//nolint:errorlint
	if he, ok := incomingError.(*echo.HTTPError); ok {
		status = he.Code
		cause = incomingError
	} else {
		switch {
		case errors.Is(incomingError, dto.ErrEntityNotFound):
			status = http.StatusNotFound
			cause = dto.ErrEntityNotFound
		default:
			status = http.StatusInternalServerError
			cause = incomingError
		}
	}
	if status != 0 {
		if !errors.As(incomingError, &errorDTO) {
			errorDTO = &dto.ErrorDTO{
				Cause:    cause,
				Message:  cause.Error(),
				TechInfo: incomingError.Error(),
			}
		}
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(status)
		} else {
			c.Response().Header().Set("content-type", echo.MIMEApplicationJSON)
			err = c.JSON(status, errorDTO)
		}
		if err != nil {
			log.Warn().Err(incomingError).Msg("Unable to create JSON response for an incoming error")
		}
		l := log.
			With().
			Int("status", status).
			Str("method", c.Request().Method).
			Str("URI", c.Request().RequestURI).
			Logger()
		l.Info().Msg(errorDTO.Message)
	}
}
