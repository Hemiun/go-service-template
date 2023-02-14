package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	"go-service-template/internal/app/infrastructure"
)

// DefaultMessageHandler is trivial kafka message handler. Just writes topic and message to the log.
func DefaultMessageHandler(ctx context.Context, message sarama.ConsumerMessage) error {
	l := infrastructure.GetBaseLogger(ctx).
		With().
		Str("topic", message.Topic).
		Str("message", string(message.Value)).
		Logger()
	l.Info().Msg("Processed")
	return nil
}
