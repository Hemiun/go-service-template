package repository

import (
	"context"

	"go-service-template/internal/app/infrastructure"
	"go-service-template/internal/app/repository/basekafkahandler"
)

// PingKafkaRepository struct for repository
type PingKafkaRepository struct {
	infrastructure.SugarLogger
	handler basekafkahandler.KafkaPinger
}

// NewPingKafkaRepository returns new PingKafkaRepository
func NewPingKafkaRepository(pinger basekafkahandler.KafkaPinger) (*PingKafkaRepository, error) {
	var target PingKafkaRepository
	target.handler = pinger
	return &target, nil
}

// Ping checks if kafka broker available or not
func (r *PingKafkaRepository) Ping(ctx context.Context, _ int) (int, error) {
	// Establishing connection and get metadata from cluster (broker list). If list is not empty then cluster is healthy
	// https://github.com/Shopify/sarama/issues/1341
	err := r.handler.Ping(ctx)
	if err != nil {
		r.LogError(ctx, "Kafka broker is unavailable", err)
		return 0, err
	}
	return 0, nil
}
