package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/assert"
	"go-service-template/internal/app/infrastructure"
)

func TestIntegrationConsumerMessage_ProduceAndConsume(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}
	var (
		consumeRes chan sarama.ConsumerMessage
		key        string
	)

	type args struct {
		timeout time.Duration
		topic   string
		message []byte
		headers map[string][]byte
	}
	tests := []struct {
		name    string
		args    args
		handler MessageHandleFunc
		wantErr bool
	}{
		{
			name: "ProduceAndConsume Case#1. Check message delivery",
			args: args{
				timeout: time.Second * 15,
				topic:   "territory.all.health-check",
				message: []byte("test message"),
				headers: map[string][]byte{"key1": []byte("val1"), "key2": []byte("val2")},
			},
			handler: func(ctx context.Context, message sarama.ConsumerMessage) error {
				l := infrastructure.GetBaseLogger(ctx).
					With().
					Str("topic", message.Topic).
					Str("message", string(message.Value)).
					Logger()
				l.Info().Msg("Processed")

				if string(message.Key) == key {
					consumeRes <- message
				}
				return nil
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.args.timeout)
			defer cancel()

			consumeRes = make(chan sarama.ConsumerMessage)
			consumer, err := NewConsumer(ctx, serviceName, kafkaConfig, postgresqlHandlerTX)
			assert.NoError(t, err)

			err = consumer.AddHandler(ctx, tt.args.topic, tt.handler)
			assert.NoError(t, err)

			go func() {
				err := consumer.Start(ctx)
				if err != nil {
					log.Error().Msg("can't start consumer service")
				}
			}()

			key = infrastructure.GenerateID()
			fl := true
			for fl {
				select {
				case <-ctx.Done():
					assert.FailNow(t, "context expired")
				default:
					if consumer.Ready() {
						fl = false
					}
					time.Sleep(time.Second)
				}
			}

			err = messageProducer.SendMessage(ctx, tt.args.topic, key, tt.args.headers, tt.args.message)
			assert.NoError(t, err)

			var incomingMsg sarama.ConsumerMessage
			select {
			case <-ctx.Done():
				assert.FailNow(t, "context expired")
			case incomingMsg = <-consumeRes:
				assert.Equal(t, tt.args.message, incomingMsg.Value)

				for k, v := range tt.args.headers {
					assert.Contains(t, incomingMsg.Headers, &sarama.RecordHeader{[]byte(k), []byte(v)}, "received headers doesnt contains value  header(%s, %s)", k, string(v))
				}

			}
		})
	}
}

func TestIntegrationConsumerMessage_Context(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	var (
		consumeRes chan sarama.ConsumerMessage
		key        string
	)

	type args struct {
		timeout time.Duration
		message []byte
		headers map[string][]byte
	}
	tests := []struct {
		name    string
		args    args
		handler MessageHandleFunc
		wantErr bool
	}{
		{
			name: "ProduceAndConsume Case#1. Check requestID",
			args: args{
				timeout: time.Minute * 5,
				message: []byte("test message"),
				headers: map[string][]byte{"key1": []byte("val1"), "key2": []byte("val2")},
			},
			handler: func(ctx context.Context, message sarama.ConsumerMessage) error {
				l := infrastructure.GetBaseLogger(ctx).
					With().
					Str("topic", message.Topic).
					Str("message", string(message.Value)).
					Logger()
				l.Info().Msg("Processed")

				if string(message.Key) == key {
					consumeRes <- message
				}
				return nil
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestID := infrastructure.GenerateID()
			ctx := context.WithValue(context.Background(), infrastructure.CtxKeyRequestID{}, requestID)
			ctx, cancel := context.WithTimeout(ctx, tt.args.timeout)
			defer cancel()

			consumeRes = make(chan sarama.ConsumerMessage)
			consumer, err := NewConsumer(ctx, serviceName, kafkaConfig, postgresqlHandlerTX)
			assert.NoError(t, err)

			err = consumer.AddHandler(ctx, "territory.all.health-check", tt.handler)
			assert.NoError(t, err)

			go func() {
				err := consumer.Start(ctx)
				if err != nil {
					log.Error().Msg("can't start consumer service")
				}
			}()

			key = infrastructure.GenerateID()
			fl := true
			for fl {
				select {
				case <-ctx.Done():
					assert.FailNow(t, "context expired")
				default:
					if consumer.Ready() {
						fl = false
					}
					time.Sleep(time.Second)
				}
			}

			err = messageProducer.SendMessage(ctx, "territory.all.health-check", key, tt.args.headers, tt.args.message)
			assert.NoError(t, err)

			var incomingMsg sarama.ConsumerMessage
			select {
			case <-ctx.Done():
				assert.FailNow(t, "context expired")
			case incomingMsg = <-consumeRes:
				for _, h := range incomingMsg.Headers {
					if string(h.Key) == infrastructure.RequestIDHeader {
						assert.Equal(t, requestID, string(h.Value))
					}
				}
			}
		})
	}
}

func TestIntegrationConsumerMessage_errorHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	type args struct {
		message     *sarama.ConsumerMessage
		occurredErr error
		expectedErr string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test_consumerHandler_errorHandler Case#1. Success",
			args: args{
				message: &sarama.ConsumerMessage{
					Headers:        []*sarama.RecordHeader{{Key: []byte("key"), Value: []byte("val")}},
					Timestamp:      time.Now().Add(-time.Second * 15),
					BlockTimestamp: time.Now().Add(-time.Second * 55),

					Key:   sarama.ByteEncoder("test_key"),
					Value: sarama.ByteEncoder("test_val"),

					Topic:     "test_topic",
					Partition: 3,
					Offset:    2,
				},
				occurredErr: errors.New("test_error_text"),
			},
			wantErr: false,
		},
		{
			name: "Test_consumerHandler_errorHandler Case#2. db error",
			args: args{
				message:     &sarama.ConsumerMessage{},
				occurredErr: errors.New("test_error_text"),
				expectedErr: pgerrcode.NotNullViolation,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messageConsumer := &consumerHandler{
				SugarLogger: infrastructure.SugarLogger{},
				handlers:    map[string]MessageHandleFunc{},
				middleware:  []MiddlewareFunc{},
				ready:       nil,
				db:          postgresqlHandlerTX,
			}

			var actualRecords sarama.ConsumerMessage
			var actualErrorText string

			// Выполняем запрос
			id, err := messageConsumer.errorHandler(context.Background(), tt.args.message, tt.args.occurredErr)

			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if tt.wantErr {
					assert.Equal(t, tt.args.expectedErr, pgErr.Code)
					return
				}
			}

			if (err != nil) != tt.wantErr {
				assert.FailNow(t, "unexpected error while errorHandler", err)
			}

			// Получаем записанные значения
			rows, err := postgresqlHandlerTX.Query(
				context.Background(),
				`SELECT 	headers_cs,
        							headers_txt,
									timestamp_cs,
									block_timestamp_cs,									
									key_cs,
									value_cs,									
									topic_cs,
									partition_cs,
									offset_cs,									
									error_text
						FROM kafka_in_error_messages WHERE id=$1`,
				id,
			)
			if err != nil {
				assert.FailNow(t, "can't get data from db", err)
			}
			flRows := false
			var tmpKey []byte
			var tmpValue []byte
			var headersText []byte
			for rows.Next() {
				flRows = true
				err = rows.Scan(
					&actualRecords.Headers,
					&headersText,
					&actualRecords.Timestamp,
					&actualRecords.BlockTimestamp,

					&tmpKey,
					&tmpValue,

					&actualRecords.Topic,
					&actualRecords.Partition,
					&actualRecords.Offset,

					&actualErrorText,
				)
				if err != nil {
					assert.FailNow(t, "error while result read", err)
				}

				actualRecords.Key = sarama.ByteEncoder(tmpKey)
				actualRecords.Value = sarama.ByteEncoder(tmpValue)
				break
			}
			if !flRows {
				assert.FailNow(t, "empty response", err)
			}

			assert.Equal(t, tt.args.message.Value, actualRecords.Value, "message.Value")
			assert.Equal(t, tt.args.message.Key, actualRecords.Key, "message.Key")
			assert.Equal(t, tt.args.message.Topic, actualRecords.Topic, "message.Topic")
			assert.Equal(t, tt.args.message.Headers, actualRecords.Headers, "message.Headers")
			assert.NotEmpty(t, headersText, "headersText")
			assert.WithinDuration(t, tt.args.message.Timestamp, actualRecords.Timestamp, time.Millisecond, "message.Timestamp")
			assert.WithinDuration(t, tt.args.message.BlockTimestamp, actualRecords.BlockTimestamp, time.Millisecond, "message.BlockTimestamp")
			assert.Equal(t, tt.args.message.Offset, actualRecords.Offset, "message.Offset")
			assert.Equal(t, tt.args.message.Partition, actualRecords.Partition, "message.Partition")

			assert.Equal(t, tt.args.occurredErr.Error(), actualErrorText, "ErrorText")
		})
	}
}

func TestIntegrationConsumerMessage_headersToJSON(t *testing.T) {
	type someType struct {
		F1 string
		F2 int
		M  map[string]any
	}

	someObj := someType{F1: "some string", F2: 112233, M: map[string]any{"key1": "val1", "key2": 4}}
	someObjB, _ := json.Marshal(someObj)

	type args struct {
		obj        []*sarama.RecordHeader
		resultJSON string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test_consumerHandler_errorHandler Case#1. Success",
			args: args{
				obj: []*sarama.RecordHeader{
					{Key: []byte("key"), Value: []byte("val")},
					{Key: []byte("objKey"), Value: someObjB},
				},
				resultJSON: "[{\"Key\":\"key\",\"Value\":\"val\"},{\"Key\":\"objKey\",\"Value\":\"{\\\"F1\\\":\\\"some string\\\",\\\"F2\\\":112233,\\\"M\\\":{\\\"key1\\\":\\\"val1\\\",\\\"key2\\\":4}}\"}]",
			},
			wantErr: false,
		},
	}
	target := consumerHandler{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := target.headersToJSON(context.Background(), tt.args.obj)
			assert.JSONEq(t, tt.args.resultJSON, string(res))
			fmt.Println(res)
		})
	}
}
