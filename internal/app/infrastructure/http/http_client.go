package http

import (
	"time"

	"github.com/go-resty/resty/v2"
)

var baseHTTPClient *resty.Client

// HTTPClientConfig - struct for HTTP client params
type HTTPClientConfig struct { //nolint:revive
	RequestTimeout time.Duration `yaml:"requestTimeout"`
}

// GetBaseHTTPClient - return prepared http client (*resty.Client)
func GetBaseHTTPClient() *resty.Client {
	return baseHTTPClient
}

// InitBaseHTTPClient - base HTTP client initialisation
func InitBaseHTTPClient(httpClient HTTPClientConfig) {
	baseHTTPClient = resty.New()

	baseHTTPClient.OnBeforeRequest(addHeaders)
	baseHTTPClient.OnBeforeRequest(logRequest)

	baseHTTPClient.OnAfterResponse(errorHandler)
	baseHTTPClient.OnAfterResponse(logRequestResult)

	if httpClient.RequestTimeout <= 0 {
		httpClient.RequestTimeout = 30 * time.Second
	}
	baseHTTPClient.SetTimeout(httpClient.RequestTimeout)
}
