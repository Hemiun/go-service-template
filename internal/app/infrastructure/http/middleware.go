package http

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-resty/resty/v2"
	"go-service-template/internal/app/infrastructure"
)

func printHeaders(header http.Header) string {
	var res string
	for key, val := range header {
		res += fmt.Sprintf("%s : %s; ", key, val)
	}
	return res
}

// logRequest - middleware for outgoing  requests logging
func logRequest(_ *resty.Client, r *resty.Request) error {
	log := infrastructure.GetBaseLogger(r.Context()).
		With().
		Str("URL", r.URL).
		Str("method", r.Method).
		Str("headers", printHeaders(r.Header)).
		Logger()
	log.Info().Msg("outgoing http request")
	return nil
}

// addHeaders - middleware for adding headers into request
func addHeaders(_ *resty.Client, r *resty.Request) error {
	requestID, ok := r.Context().Value(infrastructure.CtxKeyRequestID{}).(string)
	if !ok {
		log := infrastructure.GetBaseLogger(r.Context())
		log.Debug().Caller().Msg("request_id not found in context. new one  will be generated")
		requestID = infrastructure.GenerateID()
	}
	r.Header.Add(infrastructure.RequestIDHeader, requestID)

	correlationID, ok := r.Context().Value(infrastructure.CtxKeyCorrelationID{}).(string)
	if !ok {
		log := infrastructure.GetBaseLogger(r.Context())
		log.Debug().Caller().Msg("correlationID not found in context")
	}
	r.Header.Add(infrastructure.CorrelationIDHeader, correlationID)
	return nil
}

// logRequestResult - middleware for request result logging
func logRequestResult(_ *resty.Client, r *resty.Response) error {
	log := infrastructure.GetBaseLogger(r.Request.Context()).
		With().
		Str("URL", r.Request.URL).
		Str("method", r.Request.Method).
		Str("status", r.Status()).
		Dur("latency", r.Time()).
		Logger()
	log.Info().Send()
	return nil
}

// errorHandler - middleware for error handling
func errorHandler(_ *resty.Client, r *resty.Response) error {
	log := infrastructure.GetBaseLogger(r.Request.Context())

	var (
		err error
		ok  bool
	)
	if r.Error() != nil {
		err, ok = r.Error().(error)
		if !ok {
			return err
		}
	}

	// Обрабатываем транспортные ошибки
	if os.IsTimeout(err) {
		log.Error().Err(err).Msg("can't execute http query. client timeout occurred")
		return ErrHTTPRequestTimeout
	}
	if err != nil {
		log.Error().Err(err).Msg("can't execute http query. unexpected error")
		return err
	}

	if r.IsSuccess() && r.Request.Header.Get("accept") != r.Header().Get("content-type") {
		log.Error().Err(err).Msg("received bad content type")
		return ErrHTTPBadContentType
	}

	return nil
}
