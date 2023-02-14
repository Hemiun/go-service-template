package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-service-template/internal/app/infrastructure"

	"github.com/stretchr/testify/assert"

	httpClient "go-service-template/internal/app/infrastructure/http"

	"go-service-template/internal/app/dto"
)

func TestPingClient_Ping(t *testing.T) {
	type args struct {
		wantErr bool
		err     error
		h       http.Handler
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Ping Client. Case #1. Positive",
			args: args{
				wantErr: false,
				err:     nil,
				h: http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						responseBody, err := json.Marshal(dto.Ping{Status: "OK"})
						if err != nil {
							panic("bad dto param")
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write(responseBody)
					}),
			},
		},
		{
			name: "Ping Client. Case #2. Internal Error",
			args: args{
				wantErr: true,
				err:     httpClient.ErrHTTPRequestFailed,
				h: http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusInternalServerError)
					}),
			},
		},
		{
			name: "Ping Client. Case #3. Bad content type",
			args: args{
				wantErr: true,
				err:     httpClient.ErrHTTPBadContentType,
				h: http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						responseBody, err := json.Marshal(dto.Ping{Status: "OK"})
						if err != nil {
							panic("bad dto param")
						}
						w.Header().Set("Content-Type", "some/some")
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write(responseBody)
					}),
			},
		},
		{
			name: "Ping Client. Case #4. Timeout",
			args: args{
				wantErr: true,
				err:     httpClient.ErrHTTPRequestTimeout,
				h: http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						delayInSec := time.Second * 10
						time.Sleep(delayInSec)
						responseBody, err := json.Marshal(dto.Ping{Status: "OK"})
						if err != nil {
							panic("bad dto param")
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write(responseBody)
					}),
			},
		},
		{
			name: "Ping Client. Case #5. Bad content",
			args: args{
				wantErr: true,
				err:     httpClient.ErrHTTPBadContent,
				h: http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						responseBody := []byte("some bad content")
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write(responseBody)
					}),
			},
		},
	}
	requestID := infrastructure.GenerateID()
	ctx := context.WithValue(context.Background(), infrastructure.CtxKeyRequestID{}, requestID)

	lc := infrastructure.GetBaseLogger(context.Background()).With()
	lc = lc.Str(infrastructure.RequestIDField, requestID)
	l := lc.Logger()
	ctx = context.WithValue(ctx, infrastructure.CtxKeyLogger{}, &l)
	// для теста таймаут 5 секунд
	httpClient.InitBaseHTTPClient(httpClient.HTTPClientConfig{RequestTimeout: time.Second * 5})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.args.h)
			defer srv.Close()

			c := &PingClient{}
			c.address = srv.URL
			got, err := c.Ping(ctx)

			if tt.args.wantErr {
				assert.Equal(t, err, tt.args.err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
