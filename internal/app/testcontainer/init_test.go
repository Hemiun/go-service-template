package testcontainer

import (
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"

	"go-service-template/internal/app/infrastructure"
)

func TestMain(m *testing.M) {
	infrastructure.InitGlobalLogger("debug", "go-service-template", "")
	testcontainers.Logger = &Logger{}
	os.Exit(m.Run())
}
