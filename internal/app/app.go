package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go-service-template/internal/app/rest"

	"github.com/labstack/echo/v4"
	cfg "go-service-template/internal/app/config"
	"go-service-template/internal/app/handler"
	"go-service-template/internal/app/infrastructure"
	httpClient "go-service-template/internal/app/infrastructure/http"
	"go-service-template/internal/app/infrastructure/kafka"
	"go-service-template/internal/app/infrastructure/postgres"
	"go-service-template/internal/app/repository"
	"go-service-template/internal/app/service"
)

var (
	appConfig           *cfg.AppConfig
	dbHandler           *postgres.PostgresqlHandlerTX
	producer            *kafka.MessageProducer
	pingDBRepository    service.PingRepository
	pingKafkaRepository service.PingRepository
	pingService         handler.PingService
	pingHandler         *handler.PingHandler
	pingClient          *rest.PingClient
	resources           []Resource
	e                   *echo.Echo
	logger              *infrastructure.Logger

	consumer kafka.MessageConsumer
)

const (
	attemptsCount   = 10
	attemptInterval = time.Second * 20
)

// Resource interface used for gracefully shutdown
type Resource interface {
	Init(ctx context.Context) error
	Close(ctx context.Context) error
}

// initDatabase  initialize new dbHandler. If database is not available then "attemptInterval" attempts will be made
func initDatabase(ctx context.Context) {
	var err error
	for ind := 0; ind < attemptsCount; ind++ {
		select {
		case <-ctx.Done():
			logger.Info().Msg("database initialization aborted due to canceled context")
			return
		default:
			logger.Info().Msgf("try to initialize database: attempt #%d", ind+1)
			dataSource := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", appConfig.Database.Username, appConfig.Database.Password, appConfig.Database.Address, appConfig.Database.Port, appConfig.Database.DB) //nolint:nosprintfhostport
			dbHandler, err = postgres.NewPostgresqlHandlerTX(ctx, dataSource, appConfig.PgPool)
			if err != nil {
				logger.Error().Err(err).Msg("can't create postgres handler")
			} else {

				resources = append(resources, dbHandler)
				logger.Info().Msg("db successfully initialized")
				return
			}
			time.Sleep(attemptInterval)
		}
	}
	if dbHandler == nil {
		logger.Fatal().Err(err).Msgf("can't create postgres handler after %d attempts", attemptsCount)
	}
}

// PrepareApp - init app
func PrepareApp(ctx context.Context) {
	var err error
	appConfig = cfg.NewConfig()
	// 1. Configuration
	err = appConfig.Init()
	if err != nil {
		panic(err)
	}

	// 2. Logger
	infrastructure.InitGlobalLogger(appConfig.Logger.LogLevel, appConfig.Passport.ServiceName, appConfig.Passport.ServiceInstance)
	logger = infrastructure.GetBaseLogger(ctx)

	// 3. Init db
	initDatabase(ctx)
	pingDBRepository = repository.NewPingRepository(dbHandler)
	//

	// 4. Init producer
	initProducer(ctx)

	// 5. Init consumer
	initConsumer(ctx)

	// 6. Init kafka repository
	pingKafkaRepository, err = repository.NewPingKafkaRepository(producer)
	if err != nil {
		logger.Error().Err(err).Msg("can't create PingKafkaRepository")
	}
	//
	// 7. Http client
	httpClient.InitBaseHTTPClient(appConfig.HTTPClient)
	pingClient = rest.NewPingClient("http://localhost:8080/api/v1/ping")

	// 8. Services
	pingService = service.NewPingService(pingDBRepository, pingKafkaRepository, dbHandler)
	//

	// 9. Handler
	pingHandler = handler.NewPingHandler(pingService, pingClient)
}

// StartApp - start app
func StartApp(ctx context.Context) {
	// 1. Start echo server
	// Prepare
	prepareEcho()
	// Routes registration.
	prepareRoutes()
	// run listening
	go func() {
		if err := e.Start(appConfig.Server.Address); err != nil && errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Msgf("shutting down the server:%v", err)
			ctx.Done()
		}
	}()
	//

	// 2. Run consumer.
	// Add handler for incoming messages processing
	prepareConsumerHandlers(ctx)

	go func() {
		err := consumer.Start(ctx)
		if err != nil {
			logger.Error().Msg("can't start consumer service")
		}
	}()
	//
}

// ShutdownApp - stops processing for incoming requests and free resources
func ShutdownApp(ctx context.Context) {
	// Shutdown echo server
	if err := e.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("can't stop echo")
	}

	for _, r := range resources {
		if err := r.Close(ctx); err != nil {
			logger.Fatal().Err(err).Msg("can't free resource")
		}
	}
}
