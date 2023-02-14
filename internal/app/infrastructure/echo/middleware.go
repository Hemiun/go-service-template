package mymiddleware

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"go-service-template/internal/app/infrastructure"
)

// RequestLogger - middleware for incoming request logging
func RequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		req := c.Request()
		l := infrastructure.GetBaseLogger(c.Request().Context()).
			With().
			Str("remoteIP", req.RemoteAddr).
			Str("host", req.Host).
			Str("method", req.Method).
			Str("URI", req.RequestURI).
			Str("userAgent", req.UserAgent()).
			Logger()
		l.Info().Send()
		res := next(c)
		if res == nil {
			resp := c.Response()
			latency := time.Since(start)
			l = infrastructure.GetBaseLogger(c.Request().Context()).
				With().
				Int("status", resp.Status).
				Dur("latency", latency).
				Logger()
			l.Info().Send()
		}
		return res
	}
}

// PrepareLogger - middleware for logger configuration. Adds requestID, correlationID fields
func PrepareLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		lc := infrastructure.GetBaseLogger(context.Background()).With()
		requestID, ok := c.Request().Context().Value(infrastructure.CtxKeyRequestID{}).(string)
		if !ok {
			requestID = ""
		}

		correlationID, ok := c.Request().Context().Value(infrastructure.CtxKeyCorrelationID{}).(string)
		if !ok {
			correlationID = ""
		}
		lc = lc.Str(infrastructure.CorrelationIDField, correlationID)
		lc = lc.Str(infrastructure.RequestIDField, requestID)
		l := lc.Logger()
		newCtx := context.WithValue(c.Request().Context(), infrastructure.CtxKeyLogger{}, &l)

		newReq := c.Request().WithContext(newCtx)
		c.SetRequest(newReq)
		return next(c)
	}
}

// PrepareRequestID - adds requestID into context
func PrepareRequestID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := c.Request().Header.Get(infrastructure.RequestIDHeader)
		if requestID == "" {
			log := infrastructure.GetBaseLogger(context.Background())
			log.Debug().Caller().Msg("request_id not found in context. new one  will be generated")
			requestID = infrastructure.GenerateID()
		}
		newCtx := context.WithValue(c.Request().Context(), infrastructure.CtxKeyRequestID{}, requestID)
		newReq := c.Request().WithContext(newCtx)
		c.SetRequest(newReq)
		return next(c)
	}
}

// PrepareCorrelationID - adds correlationID into context
func PrepareCorrelationID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		correlationID := c.Request().Header.Get(infrastructure.CorrelationIDHeader)
		if correlationID == "" {
			log := infrastructure.GetBaseLogger(context.Background())
			log.Debug().Caller().Msg("correlationID not found in context")
		}
		newCtx := context.WithValue(c.Request().Context(), infrastructure.CtxKeyCorrelationID{}, correlationID)
		newReq := c.Request().WithContext(newCtx)
		c.SetRequest(newReq)
		return next(c)
	}
}
