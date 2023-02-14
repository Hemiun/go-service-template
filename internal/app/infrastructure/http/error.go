package http

import "errors"

var (
	// ErrHTTPRequestFailed - "can't execute http request" error
	ErrHTTPRequestFailed = errors.New("can't execute http request")

	// ErrHTTPRequestTimeout - "http request timeout" error
	ErrHTTPRequestTimeout = errors.New("http request timeout")
	// ErrHTTPBadContentType - "received bad content type" error
	ErrHTTPBadContentType = errors.New("received bad content type")
	// ErrHTTPBadContent - "received bad type" error
	ErrHTTPBadContent = errors.New("received bad type")
)
