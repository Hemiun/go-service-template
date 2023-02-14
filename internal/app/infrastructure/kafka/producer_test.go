package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-service-template/internal/app/infrastructure"

	"github.com/Shopify/sarama"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestIntegrationMessageProducer_SendMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	type args struct {
		topic   string
		key     string
		value   string
		headers map[string][]byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "MessageProducer.SendMessage Case #1. Just send to topic",
			args: args{
				topic:   "territory.all.test-topic",
				key:     "test_key",
				value:   "test message",
				headers: map[string][]byte{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := context.Background()
			c = context.WithValue(c, infrastructure.CtxKeyCorrelationID{}, "CorrelationIDValue")
			require.NoError(t, messageProducer.SendMessage(c, tt.args.topic, tt.args.key, tt.args.headers, []byte(tt.args.value)))
		})
	}
}

func TestIntegrationMessageProducer_Ping(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	type args struct {
		topic string
		key   string
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "MessageProducer.Ping Case #1.",
			args: args{
				topic: "territory.all.test-topic",
				key:   "test_key",
				value: "test message",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, messageProducer.Ping(context.Background()))
		})
	}
}

func TestIntegrationMessageProducer_errorHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	type args struct {
		message     *sarama.ProducerMessage
		occurredErr error
		expectedErr string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestMessageProducer_errorHandler Case#1. Success",
			args: args{
				message: &sarama.ProducerMessage{
					Topic:     "test_topic",
					Key:       sarama.StringEncoder("test_key"),
					Value:     sarama.ByteEncoder("test_val"),
					Headers:   []sarama.RecordHeader{{Key: []byte("key"), Value: []byte("val")}},
					Metadata:  map[string]any{"test": "val"},
					Offset:    2,
					Partition: 3,
					Timestamp: time.Now().Add(-time.Second * 15),
				},
				occurredErr: errors.New("test_error_text"),
			},

			wantErr: false,
		},
		{
			name: "TestMessageProducer_errorHandler Case#2. db error - value = null",
			args: args{
				message: &sarama.ProducerMessage{
					Topic: "test_topic",
					Key:   sarama.StringEncoder("test_key"),
					// Value:     sarama.ByteEncoder("test_val"),
					Headers:   []sarama.RecordHeader{{Key: []byte("key"), Value: []byte("val")}},
					Metadata:  map[string]any{"test": "val"},
					Offset:    2,
					Partition: 3,
					Timestamp: time.Now().Add(-time.Second * 15),
				},
				occurredErr: errors.New("test_error_text"),
				expectedErr: pgerrcode.NotNullViolation,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actualRecords sarama.ProducerMessage
			var actualErrorText string

			id, err := messageProducer.errorHandler(context.Background(), tt.args.message, tt.args.occurredErr)

			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if tt.wantErr {
					assert.Equal(t, tt.args.expectedErr, pgErr.Code)
					return
				}
			}

			if (err != nil) != tt.wantErr {
				assert.FailNow(t, "unexpected error", err)
			}

			rows, err := postgresqlHandlerTX.Query(
				context.Background(),
				`SELECT 	topic_pc,
									key_pc,
									value_pc,
									headers_pc,
									headers_txt,
									metadata_pc,
									offset_pc,
									partition_pc,
									timestamp_pc,
									error_text 
						FROM kafka_out_error_messages WHERE id=$1`,
				id,
			)

			flRows := false
			var tmpKey []byte
			var tmpValue []byte
			var headersText []byte

			for rows.Next() {
				flRows = true
				err = rows.Scan(
					&actualRecords.Topic,
					&tmpKey,
					&tmpValue,
					&actualRecords.Headers,
					&headersText,
					&actualRecords.Metadata,
					&actualRecords.Offset,
					&actualRecords.Partition,
					&actualRecords.Timestamp,

					&actualErrorText,
				)
				if err != nil {
					assert.FailNow(t, "error while result read", err)
				}

				actualRecords.Key = sarama.StringEncoder(tmpKey)
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
			assert.NotEmpty(t, headersText)
			assert.WithinDuration(t, tt.args.message.Timestamp, actualRecords.Timestamp, time.Millisecond, "message.Timestamp")
			assert.Equal(t, tt.args.message.Offset, actualRecords.Offset, "message.Offset")
			assert.Equal(t, tt.args.message.Partition, actualRecords.Partition, "message.Partition")

			assert.Equal(t, tt.args.occurredErr.Error(), actualErrorText, "ErrorText")
		})
	}
}
