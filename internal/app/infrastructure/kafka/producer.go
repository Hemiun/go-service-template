package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Shopify/sarama"
	"go-service-template/internal/app/infrastructure"
)

const (
	sequenceNextIDOutErrorMessage = "kafka_out_error_messages_sq"

	addKafkaOutErrorMessage = `INSERT INTO kafka_out_error_messages
	(	id,
		topic_pc,
		key_pc,
		value_pc,
		headers_pc,
	    headers_txt,
		metadata_pc,
		offset_pc,
		partition_pc,
		timestamp_pc,
		error_text,
		send_time)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
)

// ErrNoKafkaBrokers - "can't get kafka broker" error
var ErrNoKafkaBrokers = errors.New("can't get kafka broker")

// MessageProducer struct for interactions with kafka cluster
type MessageProducer struct {
	infrastructure.SugarLogger
	brokers  []string
	config   *sarama.Config
	producer sarama.SyncProducer
	db       db
}

// NewMessageProducer return new MessageProducer
func NewMessageProducer(ctx context.Context, kafkaConfig KafkaConfig, db db) (*MessageProducer, error) {
	var target MessageProducer
	if kafkaConfig.LogSarama {
		sarama.Logger = infrastructure.GetSaramaLogger(ctx)
	}

	target.brokers = kafkaConfig.BrokerList
	target.db = db

	if err := target.Init(ctx); err != nil {
		return nil, err
	}

	return &target, nil
}

// Init - func for initialisation MessageProducer
func (h *MessageProducer) Init(ctx context.Context) error {
	h.config = sarama.NewConfig()
	h.config.Producer.Partitioner = sarama.NewRandomPartitioner
	h.config.Producer.RequiredAcks = sarama.WaitForAll
	h.config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(h.brokers, h.config)
	if err != nil {
		h.LogError(ctx, "Can't get new sync messageProducer", err)
		return err
	}
	h.producer = producer

	return nil
}

// SendMessage - send message into the kafka topic
func (h *MessageProducer) SendMessage(ctx context.Context, topic string, key string, headers map[string][]byte, message []byte) error {
	saramaRecordHeaders := make([]sarama.RecordHeader, 0)

	if _, ok := headers[infrastructure.RequestIDHeader]; !ok {
		requestID, ok := ctx.Value(infrastructure.CtxKeyRequestID{}).(string)
		if !ok {
			h.LogWarn(ctx, "requestID not set in context")
		}
		headers[infrastructure.RequestIDHeader] = []byte(requestID)
	}

	if _, ok := headers[infrastructure.CorrelationIDHeader]; !ok {
		correlationID, ok := ctx.Value(infrastructure.CtxKeyCorrelationID{}).(string)
		if ok {
			headers[infrastructure.CorrelationIDHeader] = []byte(correlationID)
		}
	}

	for headerKey, headerValue := range headers {
		saramaRecordHeaders = append(saramaRecordHeaders, sarama.RecordHeader{Key: []byte(headerKey), Value: headerValue})
	}

	producerMessage := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.StringEncoder(key),
		Value:   sarama.ByteEncoder(message),
		Headers: saramaRecordHeaders,
	}
	partition, offset, err := h.producer.SendMessage(producerMessage)
	if err != nil {
		h.LogError(ctx, "Can' send message to kafka topic", err)
		// errorHandler must be started with new context!
		_, err2 := h.errorHandler(context.Background(), producerMessage, err) //nolint:contextcheck
		if err2 != nil {
			h.LogError(ctx, "db error", err2)
			return err2
		}
		return nil
	}
	log := infrastructure.GetBaseLogger(ctx).
		With().
		Str("topic", topic).
		Str("key", key).
		Int32("partition", partition).
		Int64("offset", offset).
		Logger()
	log.Info().Msg("message send to kafka")
	return nil
}

// Close - close allocated resources
func (h *MessageProducer) Close(ctx context.Context) error {
	err := h.producer.Close()
	if err != nil {
		h.LogError(ctx, "Can't close messageProducer", err)
	}

	return err
}

// Ping -  check is kafka broker available or not
func (h *MessageProducer) Ping(_ context.Context) error {
	prodClient, err := sarama.NewClient(h.brokers, h.config)
	if err != nil {
		return err
	}
	defer prodClient.Close() //nolint:errcheck
	b := prodClient.Brokers()
	if len(b) < 1 {
		return ErrNoKafkaBrokers
	}
	return nil
}

func (h *MessageProducer) errorHandler(ctx context.Context, message *sarama.ProducerMessage, occurredErr error) (int64, error) {
	id, err := h.db.GetNextID(ctx, sequenceNextIDOutErrorMessage)
	if err != nil {
		return 0, err
	}
	//

	err = h.db.Execute(ctx, addKafkaOutErrorMessage,
		id,

		message.Topic,
		message.Key,
		message.Value,
		message.Headers,
		h.headersToJSON(ctx, message.Headers),
		message.Metadata,
		message.Offset,
		message.Partition,
		message.Timestamp,

		occurredErr.Error(),
		time.Now())
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (h *MessageProducer) headersToJSON(ctx context.Context, obj []sarama.RecordHeader) []byte {
	type rec struct{ Key, Value string }
	buf := make([]rec, len(obj))
	for ind, el := range obj {
		buf[ind] = rec{Key: string(el.Key), Value: string(el.Value)}
	}
	res, err := json.Marshal(buf)
	if err != nil {
		h.LogWarn(ctx, "can't marshall header struct")
	}
	return res
}
