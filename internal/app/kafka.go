package app

import (
	"context"
	"time"

	kafka2 "go-service-template/internal/app/handler/kafka"
	"go-service-template/internal/app/infrastructure/kafka"
)

func initProducer(ctx context.Context) {
	var err error

	for ind := 0; ind < attemptsCount; ind++ {
		select {
		case <-ctx.Done():
			logger.Info().Msg("kafka producer initialization aborted due to canceled context")
			return
		default:
			logger.Info().Msgf("try to initialize kafka producer: attempt #%d", ind+1)
			producer, err = kafka.NewMessageProducer(ctx, appConfig.Kafka, dbHandler)
			if err != nil {
				logger.Error().Err(err).Msg("can't create kafka message producer")
			} else {
				resources = append(resources, producer)
				return
			}
			time.Sleep(attemptInterval)
		}
	}
	if producer == nil {
		logger.Fatal().Err(err).Msgf("can't create kafka producer after %d attempts", attemptsCount)
	}
}

func initConsumer(ctx context.Context) {
	var err error
	for ind := 0; ind < attemptsCount; ind++ {
		select {
		case <-ctx.Done():
			logger.Info().Msg("kafka consumer initialization aborted due to canceled context")
			return
		default:
			logger.Info().Msgf("try to initialize kafka consumer: attempt #%d", ind+1)
			consumer, err = kafka.NewConsumer(ctx, appConfig.Passport.ServiceName, appConfig.Kafka, dbHandler)
			if err != nil {
				logger.Error().Err(err).Msg("can't create kafka consumer")
			} else {
				resources = append(resources, consumer)
				return
			}
			time.Sleep(attemptInterval)
		}
	}
	if consumer == nil {
		logger.Fatal().Err(err).Msgf("can't create kafka producer after %d attempts", attemptsCount)
	}
}

func prepareConsumerHandlers(ctx context.Context) {
	// Add new handler
	err := consumer.AddHandler(ctx, "territory.all.health-check", kafka2.DefaultMessageHandler)
	if err != nil {
		logger.Error().Msg("can't add DefaultMessageHandler to kafka consumer")
	}
}
