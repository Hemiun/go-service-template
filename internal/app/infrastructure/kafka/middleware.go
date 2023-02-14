package kafka

import (
	"context"
	"time"

	"github.com/Shopify/sarama"
	"go-service-template/internal/app/infrastructure"
)

func logIncomingMessageMiddleware(next MessageHandleFunc) MessageHandleFunc {
	return func(ctx context.Context, message sarama.ConsumerMessage) error {
		log := infrastructure.GetBaseLogger(ctx).
			With().
			Str("topic", message.Topic).
			Time("messageTimestamp", message.Timestamp).
			Int32("partition", message.Partition).
			Int64("offset", message.Offset).
			Logger()
		log.Info().Msg("Incoming message")
		start := time.Now()
		status := "success"
		res := next(ctx, message)
		if res != nil {
			status = "error"
		}
		latency := time.Since(start)
		log = infrastructure.GetBaseLogger(ctx).
			With().
			Str("topic", message.Topic).
			Time("messageTimestamp", message.Timestamp).
			Int32("partition", message.Partition).
			Int64("offset", message.Offset).
			Str("status", status).
			Dur("latency", latency).
			Logger()
		log.Info().Send()
		return res
	}
}

func prepareLoggerMiddleware(next MessageHandleFunc) MessageHandleFunc {
	return func(ctx context.Context, message sarama.ConsumerMessage) error {
		var (
			requestID     string
			correlationID string
		)
		lc := infrastructure.GetBaseLogger(context.Background()).With() //nolint:contextcheck

		for _, m := range message.Headers {
			if string(m.Key) == infrastructure.RequestIDHeader {
				requestID = string(m.Value)
			}

			if string(m.Key) == infrastructure.CorrelationIDHeader {
				correlationID = string(m.Value)
			}
		}

		if requestID == "" {
			requestID = infrastructure.GenerateID()
		}

		lc = lc.Str(infrastructure.RequestIDField, requestID)
		if correlationID != "" {
			lc = lc.Str(infrastructure.CorrelationIDField, correlationID)
		}

		l := lc.Logger()
		newCtx := context.WithValue(ctx, infrastructure.CtxKeyLogger{}, &l)
		newCtx = context.WithValue(newCtx, infrastructure.CtxKeyRequestID{}, requestID)
		newCtx = context.WithValue(newCtx, infrastructure.CtxKeyCorrelationID{}, correlationID)
		res := next(newCtx, message)
		return res
	}
}
