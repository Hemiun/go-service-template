package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "go-service-template/docs/swagger"
	"go-service-template/internal/app/infrastructure"
)

// Main - main entry point
// @title Echo Swagger  API
// @version 1.0
// @description go-service-template
// @host localhost:8080
// @BasePath /api/v1
// @schemes http
func Main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Waiting for a shutdown signal from os. When it received cancel the context
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		logger = infrastructure.GetBaseLogger(ctx)
		logger.Warn().Msgf("Signal '%s' was caught. Exiting", s)
		cancel()
	}()

	PrepareApp(ctx)

	StartApp(ctx)

	<-ctx.Done()
	shutdownCtx, cancelTerminator := context.WithTimeout(context.Background(), 60*time.Second)
	defer func() {
		cancelTerminator()
	}()
	ShutdownApp(shutdownCtx)
}
