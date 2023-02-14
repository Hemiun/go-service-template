package kafka

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"go-service-template/internal/app/infrastructure/postgres"

	"go-service-template/internal/app/testcontainer"

	"go-service-template/internal/app/infrastructure"
)

var (
	postgresqlHandlerTX *postgres.PostgresqlHandlerTX
	messageProducer     *MessageProducer
	kafkaConfig         KafkaConfig
	log                 *infrastructure.Logger
	serviceName         string
)

func printBaner(dsn string) {
	frameSize := len(dsn) * 2
	padSize := (frameSize - len(dsn)) / 2
	pad := strings.Repeat(" ", padSize)
	fmt.Print("\n")
	fmt.Println(strings.Repeat("-", frameSize))
	fmt.Println(fmt.Sprintf("%s%s%s", pad, dsn, pad))
	fmt.Println(strings.Repeat("-", frameSize))
	fmt.Print("\n")
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	serviceName = "go-service-template"
	infrastructure.InitGlobalLogger("default", serviceName, "")
	log = infrastructure.GetBaseLogger(ctx)
	flag.Parse()

	if !testing.Short() {
		postgresConfig := testcontainer.DatabaseContainerConfig{
			DatabaseName:      "territory",
			OwnerSchema:       "test",
			OwnerSchemaPass:   "test",
			ServiceSchema:     "test_ms",
			ServiceSchemaPass: "test_ms",
			Timeout:           time.Minute,
		}

		db, err := testcontainer.NewDatabaseContainer(ctx, postgresConfig)
		if err != nil {
			log.Panic().Err(err).Msg("can't init db container")
		}
		err = db.PrepareDB(ctx)
		if err != nil {
			log.Panic().Err(err).Msg("can't prepare db")
		}
		defer db.Close(ctx)
		dsn := db.ConnectionString(ctx)
		printBaner(dsn)
		postgresqlHandlerTX, err = postgres.NewPostgresqlHandlerTX(ctx, dsn, postgres.PgPoolConfig{})
		if err != nil {
			log.Panic().Err(err).Msg("can't init PostgresqlHandlerTX")
		}

		kafkaContainer, err := testcontainer.NewKafkaContainer(ctx, testcontainer.KafkaContainerConfig{Network: "Net4Test", Timeout: time.Minute * 3})
		if err != nil {
			log.Panic().Err(err).Msg("can't init kafka container")
		}
		brokers, err := kafkaContainer.GetBrokerList(ctx)
		if err != nil {
			log.Panic().Err(err).Msg("can't get broker list")
		}
		printBaner("kafka://" + brokers[0])
		kafkaConfig = KafkaConfig{BrokerList: brokers, LogSarama: true}
		messageProducer, err = NewMessageProducer(context.Background(), kafkaConfig, postgresqlHandlerTX)
		if err != nil {
			fmt.Println("can't init kafka handler")
			os.Exit(1)
		}
	}

	os.Exit(m.Run())
}
