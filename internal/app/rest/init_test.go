package rest

import (
	"os"
	"testing"

	"go-service-template/internal/app/infrastructure"
)

func TestMain(m *testing.M) {
	infrastructure.InitGlobalLogger("Debug", "go-service-template", "")

	os.Exit(m.Run())
}
