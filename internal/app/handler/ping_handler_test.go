package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go-service-template/internal/app/dto"
	"go-service-template/internal/app/handler/mocks"
)

func TestPingHandler_Ping(t *testing.T) {
	type wants struct {
		responseCode int
		contentType  string
		json         string
	}
	type args struct {
		wantError bool
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "PingHandler.Ping Case#1 Positive",
			args: args{
				wantError: false,
			},
			wants: wants{
				responseCode: http.StatusOK,
				contentType:  echo.MIMEApplicationJSON,
				json:         `{"status": "OK", "statusDB": "OK", "statusKafka":"OK", "message":""}`,
			},
		},
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	pingService := mocks.NewMockPingService(mockCtrl)
	pingClient := mocks.NewMockPingClient(mockCtrl)
	target := NewPingHandler(pingService, pingClient)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.args.wantError {
				pingService.EXPECT().Ping(gomock.Any()).Return(dto.Ping{Status: "OK", StatusDB: "OK", StatusKafka: "OK"})
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Assertions
			if assert.NoError(t, target.PingHandler(c)) {
				assert.Equal(t, http.StatusOK, rec.Code)
				contentType := rec.Header().Get("content-type")
				assert.Equal(t, tt.wants.contentType, contentType, "Expected status %d, got %d", tt.wants.contentType, contentType)
				// assert.Equal(t, dto.Ping{Status: "OK"}, rec.Body.String())
				assert.JSONEq(t, tt.wants.json, rec.Body.String())
			}
		})
	}
}
