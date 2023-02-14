package basekafkahandler

import (
	"context"
)

const (
	// OffsetOldest - get all messages
	OffsetOldest = -2

	// OffsetNewest - get only new messages
	OffsetNewest = -1
)

// KafkaProducer interface for sending messages into kafka
type KafkaProducer interface {
	SendMessage(ctx context.Context, topic string, key string, message []byte) error
}

// KafkaPinger - interface for checking kafka broker availability
type KafkaPinger interface {
	Ping(ctx context.Context) error
}
