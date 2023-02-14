package testcontainer

import (
	"context"
	"testing"
	"time"

	"go-service-template/internal/app/infrastructure"

	"github.com/stretchr/testify/assert"

	"github.com/testcontainers/testcontainers-go"
)

const testNetwork = "Net4Test"

func TestIntegrationKafkaContainer_initNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	tests := []struct {
		name string
		cfg  KafkaContainerConfig
	}{
		{
			name: "Case 1. Positive(init Network)",
			cfg: KafkaContainerConfig{
				Timeout: time.Minute * 5,
				Network: testNetwork,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.cfg.Timeout)
			defer cancel()
			cli, _, _, err := testcontainers.NewDockerClient()
			assert.NoError(t, err)
			target := KafkaContainer{cfg: tt.cfg, dockerClient: cli, sessionID: infrastructure.GenerateID()}
			err = target.initNetwork(ctx)
			assert.NoError(t, err)
			defer func() {
				err := cli.NetworkRemove(ctx, target.networkID)
				assert.NoError(t, err)
			}()
		})
	}
}

func TestIntegrationKafkaContainer_InitZoo(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	tests := []struct {
		name string
		cfg  KafkaContainerConfig
	}{
		{
			name: "Case 1. Positive(prepare zookeeper)",
			cfg: KafkaContainerConfig{
				Timeout: time.Minute,
				Network: testNetwork,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.cfg.Timeout)
			defer cancel()
			target := KafkaContainer{cfg: tt.cfg}
			err := target.initZookeeper(ctx)
			defer func() {
				if target.zoo != nil {
					_ = target.zoo.Terminate(ctx)
				}
			}()
			// target.zoo.Terminate(ctx) //nolint:errcheck
			assert.NoError(t, err)
		})
	}
}

func TestIntegrationKafkaContainer_all(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	tests := []struct {
		name string
		cfg  KafkaContainerConfig
	}{
		{
			name: "Case 1. Positive(init all kafka)",
			cfg: KafkaContainerConfig{
				Timeout: time.Minute * 3,
				Network: testNetwork,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.cfg.Timeout)
			defer cancel()
			target, err := NewKafkaContainer(ctx, tt.cfg)
			if err != nil {
				assert.FailNowf(t, "can't init container", "%v", err)
			}
			defer target.Close(ctx)

			brokerList, err := target.GetBrokerList(ctx)
			assert.NoError(t, err)
			assert.NotEmpty(t, brokerList)
		})
	}
}

func TestIntegrationKafkaContainer_cleanNetworks(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	tests := []struct {
		name string
		cfg  KafkaContainerConfig
	}{
		{
			name: "Case 1. Positive(init all kafka)",
			cfg: KafkaContainerConfig{
				Timeout: time.Minute,
				Network: testNetwork,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.cfg.Timeout)
			defer cancel()
			cli, _, _, err := testcontainers.NewDockerClient()
			assert.NoError(t, err)
			target := KafkaContainer{cfg: tt.cfg, dockerClient: cli}
			assert.NotPanics(t, func() { target.cleanNetworks(ctx) })
		})
	}
}
