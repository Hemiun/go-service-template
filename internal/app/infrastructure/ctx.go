package infrastructure

// keys for context params
type (
	// CtxKeyLogger - key for  logger context param
	CtxKeyLogger struct{}

	// CtxKeyRequestID - key for requestID context param
	CtxKeyRequestID struct{}

	// CtxKeyTransaction - key for transaction context param
	CtxKeyTransaction struct{}

	// CtxKeyCorrelationID  - key for  correlationID context param
	CtxKeyCorrelationID struct{}
)

const (
	// RequestIDHeader - name of the header for requestID storing. Used for HTTP requests and kafka messages
	RequestIDHeader = "X-Request-Id"

	// RequestIDField - requestID  field name. Used for logger
	RequestIDField = "requestID"

	// CorrelationIDHeader - name of the header for correlationID storing. Used for HTTP requests and kafka messages
	CorrelationIDHeader = "X-Correlation-Id"

	// CorrelationIDField - correlationID field name. Used for logger
	CorrelationIDField = "correlationID"
)
