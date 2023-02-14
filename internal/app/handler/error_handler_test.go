package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go-service-template/internal/app/dto"
)

func TestErrorHandler(t *testing.T) {
	type args struct {
		incomingError error
	}
	type wants struct {
		responseCode int
		contentType  string
		json         string
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "Case 1. No error",
			args: args{incomingError: nil},
			wants: wants{
				responseCode: http.StatusOK,
				contentType:  echo.MIMEApplicationXML,
				json:         "",
			},
		},
		{
			name: "Case 2. Custom error",
			args: args{incomingError: dto.ErrEntityNotFound},
			wants: wants{
				responseCode: http.StatusNotFound,
				contentType:  echo.MIMEApplicationJSON,
				json:         `{"message":"entity not found","techInfo":"entity not found"}`,
			},
		},
		{
			name: "Case 3. HTTPError",
			args: args{incomingError: &echo.HTTPError{Code: 500, Message: "some http error"}},
			wants: wants{
				responseCode: http.StatusInternalServerError,
				contentType:  echo.MIMEApplicationJSON,
				json:         `{"message":"code=500, message=some http error", "techInfo":"code=500, message=some http error"}`,
			},
		},
	}

	e := echo.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Response().Header().Set("content-type", echo.MIMEApplicationXML)
			ErrorHandler(tt.args.incomingError, c)
			assert.Equal(t, tt.wants.responseCode, rec.Code, "Expected response is  %d, got %d", tt.wants.responseCode, rec.Code)
			contentType := rec.Header().Get("content-type")
			assert.Equal(t, tt.wants.contentType, contentType, "Expected status %d, got %d", tt.wants.contentType, contentType)
			if tt.wants.json != "" {
				assert.JSONEq(t, tt.wants.json, rec.Body.String())
			}
		})
	}
}
