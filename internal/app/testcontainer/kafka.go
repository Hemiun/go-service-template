package testcontainer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/go-playground/validator/v10"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go-service-template/internal/app/infrastructure"
)

var (
	ErrNoBrokerAvailable     = errors.New("no broker available")
	ErrConfigValidationError = errors.New("config validation error")
	ErrNoDockerClient        = errors.New("no docker client available")
)

// KafkaContainerConfig - config struct for container with kafka broker
type KafkaContainerConfig struct {
	// Timeout. After expiration context will be canceled
	Timeout time.Duration `validate:"required"`
	// Network. Prefix for network name
	Network string `validate:"required"`
	// Kafka. Prefix for kafka container name
	Kafka string
	// ZooKeeper  Prefix for zookeeper container name
	ZooKeeper string

	// Waiting. Time period for kafka launch waiting for
	Waiting time.Duration
}

// Validate - validate config struct
func (c *KafkaContainerConfig) Validate() error {
	if c.Network == "" {
		c.Network = "Net4Test"
	}

	if c.Kafka == "" {
		c.Kafka = "KafkaBroker"
	}

	if c.ZooKeeper == "" {
		c.ZooKeeper = "ZooKeeper"
	}

	if c.Waiting == 0 {
		c.Waiting = time.Minute * 2
	}

	return validator.New().Struct(c)
}

// KafkaContainer - test container for kafka broker
type KafkaContainer struct {
	infrastructure.SugarLogger
	zoo          testcontainers.Container
	broker       testcontainers.Container
	cfg          KafkaContainerConfig
	networkID    string
	sessionID    string
	networkName  string
	zooName      string
	dockerClient *client.Client
	brokerPort   string
}

// NewKafkaContainer - returns new KafkaContainer
func NewKafkaContainer(ctx context.Context, cfg KafkaContainerConfig) (*KafkaContainer, error) {
	testcontainers.Logger = &Logger{}
	var target KafkaContainer
	if err := cfg.Validate(); err != nil {
		target.LogError(ctx, "Config validation error", err)
		return nil, ErrConfigValidationError
	}
	target.cfg = cfg
	ctxWithTimeout, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	target.sessionID = infrastructure.GenerateID()

	err := target.initDockerClient()
	if err != nil {
		target.LogError(ctx, "Can't init docker client", err)
		return nil, ErrNoDockerClient
	}

	err = target.initNetwork(ctxWithTimeout)
	if err != nil {
		target.LogError(ctx, "Can't init network for containers", err)
		return nil, err
	}

	err = target.initZookeeper(ctx)
	if err != nil {
		target.LogError(ctx, "Can't init zookeeper", err)
		return nil, err
	}

	err = target.initKafkaBroker(ctx)
	if err != nil {
		target.LogError(ctx, "Can't init kafka broker", err)
		return nil, err
	}

	return &target, nil
}

// GetBrokerList - retuen endpoint list for kafka connecting
//
func (target *KafkaContainer) GetBrokerList(ctx context.Context) ([]string, error) {
	if target.broker == nil {
		return nil, ErrNoBrokerAvailable
	}
	host, err := target.broker.Host(ctx)
	if err != nil {
		target.LogError(ctx, "Can't get endpoint from kafka broker container", err)
		return nil, ErrNoBrokerAvailable
	}
	endpoint := fmt.Sprintf("%s:%s", host, target.brokerPort)
	return []string{endpoint}, nil
}

// Close - destruct all run containers
func (target *KafkaContainer) Close(ctx context.Context) {
	var err error
	if target.broker != nil {
		err = target.broker.Terminate(ctx)
		if err != nil {
			target.LogError(ctx, "Error while broker termination", err)
		}
	}
	if target.zoo != nil {
		err = target.zoo.Terminate(ctx)
		if err != nil {
			target.LogError(ctx, "Error while zoo termination", err)
		}
	}
	err = target.dockerClient.NetworkRemove(ctx, target.networkID)
	if err != nil {
		target.LogError(ctx, "Error while network termination", err)
	}
}

func (target *KafkaContainer) initNetwork(ctx context.Context) error {
	target.LogDebug(ctx, "Try to create network")
	if target.dockerClient == nil {
		target.LogError(ctx, "Docker client not initialized", nil)
		return ErrNoDockerClient
	}
	networkName := target.cfg.Network + target.sessionID
	target.LogDebug(ctx, "Network name is "+networkName)
	netFilterArgs := filters.NewArgs()
	netFilterArgs.Add("name", networkName)
	target.LogDebug(ctx, "Try to find an existing network")
	foundNet, err := target.dockerClient.NetworkList(ctx, types.NetworkListOptions{Filters: netFilterArgs})
	if err != nil {
		target.LogError(ctx, "Can't execute query with docker client", err)
		return err
	}
	if len(foundNet) == 0 {
		target.LogDebug(ctx, "Network not found. Try to create. Network name is "+networkName)
		n, err := target.dockerClient.NetworkCreate(ctx, networkName, types.NetworkCreate{CheckDuplicate: true})
		if err != nil {
			target.LogError(ctx, "Can't execute create network", err)
			return err
		}
		target.networkID = n.ID
		target.networkName = networkName
		target.LogDebug(ctx, fmt.Sprintf("Network created. Network name is %s, id is %s", networkName, n.ID))
		return nil
	}
	target.networkID = foundNet[0].ID
	target.networkName = networkName
	target.LogDebug(ctx, fmt.Sprintf("Network found. Network name is %s, id is %s", networkName, foundNet[0].ID))
	return nil
}

func (target *KafkaContainer) initZookeeper(ctx context.Context) error {
	target.LogDebug(ctx, "Try to init zookeeper")
	zooName := target.cfg.ZooKeeper + target.sessionID
	target.zooName = zooName
	req := testcontainers.ContainerRequest{
		Name:     zooName,
		Networks: []string{target.networkName},
		Image:    "confluentinc/cp-zookeeper:7.2.2",
		ExposedPorts: []string{
			"2181/tcp",
		},
		AutoRemove: true,
		Env: map[string]string{
			"ZOOKEEPER_CLIENT_PORT": "2181",
			"ZOOKEEPER_TICK_TIME":   "2000",
		},
		WaitingFor: wait.ForListeningPort("2181/tcp"),
	}
	zoo, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return err
	}
	target.zoo = zoo
	name, _ := zoo.Name(ctx)
	target.LogDebug(ctx, fmt.Sprintf("Zookeper created. DokerID:%s, Name:%s", zoo.GetContainerID(), name))
	return nil
}

// getFreePort - return free port
func (target *KafkaContainer) getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = l.Close()
	}()
	listenerAddr, _ := l.Addr().(*net.TCPAddr)
	return listenerAddr.Port, nil
}

func (target *KafkaContainer) initKafkaBroker(ctx context.Context) error {
	target.LogDebug(ctx, "Try to init kafka broker")
	brokerName := target.cfg.Kafka + target.sessionID

	port, err := target.getFreePort()
	if err != nil {
		target.LogError(ctx, "Can't get free port", err)
		return err
	}
	target.brokerPort = strconv.Itoa(port)

	expose := fmt.Sprintf("%d:9092", port)
	zooEndpoint := fmt.Sprintf("%s:%d", target.zooName, 2181)
	advertisedListeners := fmt.Sprintf("INTERNAL://localhost:29092,EXTERNAL://localhost:%d", port)

	req := testcontainers.ContainerRequest{
		Name:     brokerName,
		Networks: []string{target.networkName},
		Image:    "confluentinc/cp-server:7.2.2",
		ExposedPorts: []string{
			expose,
		},
		AutoRemove: true,
		Env: map[string]string{
			"KAFKA_BROKER_ID":                                   "1",
			"KAFKA_ZOOKEEPER_CONNECT":                           zooEndpoint,
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":              "INTERNAL:PLAINTEXT,EXTERNAL:PLAINTEXT",
			"KAFKA_ADVERTISED_LISTENERS":                        advertisedListeners,
			"KAFKA_LISTENERS":                                   "INTERNAL://:29092,EXTERNAL://0.0.0.0:9092",
			"KAFKA_INTER_BROKER_LISTENER_NAME":                  "INTERNAL",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":            "1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":            "0",
			"KAFKA_CONFLUENT_LICENSE_TOPIC_REPLICATION_FACTOR":  "1",
			"KAFKA_CONFLUENT_BALANCER_TOPIC_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":               "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR":    "1",
			"CONFLUENT_METRICS_ENABLE":                          "false",
		},
		WaitingFor: NewMetadataWaitStrategy(time.Minute * 3),
	}
	target.LogDebug(ctx, "Final broker configuration is:")
	target.LogDebug(ctx, fmt.Sprint(req.Env))
	broker, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return err
	}
	target.broker = broker
	name, _ := broker.Name(ctx)
	target.LogDebug(ctx, fmt.Sprintf("Kafka broker created. DokerID:%s, Name:%s", broker.GetContainerID(), name))
	return nil
}

func (target *KafkaContainer) initDockerClient() error {
	cli, _, _, err := testcontainers.NewDockerClient()
	target.dockerClient = cli
	if err != nil {
		return err
	}
	return nil
}

// cleanNetworks - removes docker networks if name match pattern
func (target *KafkaContainer) cleanNetworks(ctx context.Context) { //nolint:unused
	foundNets, err := target.dockerClient.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		target.LogError(ctx, "can't get network list", err)
		return
	}
	for _, n := range foundNets {
		if strings.Contains(n.Name, target.cfg.Network) {
			target.LogDebug(ctx, fmt.Sprintf("try to remove network: Name is %s, ID is %s", n.Name, n.ID))
			err := target.dockerClient.NetworkRemove(ctx, n.ID)
			if err != nil {
				target.LogError(ctx, fmt.Sprintf("can't remove network: Name is %s, ID is %s", n.Name, n.ID), err)
			}
		}
	}
}
