package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"go-service-template/internal/app/infrastructure"
)

// ErrBadParam - "bad param occurred" error
var ErrBadParam = errors.New("bad param occurred")

const (
	consumptionStopped           bool = false
	consumptionStarted           bool = true
	sequenceNextIDInErrorMessage      = "kafka_in_error_messages_sq"

	addKafkaInErrorMessage = `INSERT INTO kafka_in_error_messages
	(	id,
		headers_cs,
	    headers_txt,
		timestamp_cs,
		block_timestamp_cs,
		key_cs,
		value_cs,
		topic_cs,
		partition_cs,
		offset_cs,
		error_text,
		receive_time)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11, $12)`
)

// MessageHandleFunc  - func type for kafka message handlers
type MessageHandleFunc func(ctx context.Context, message sarama.ConsumerMessage) error

// MiddlewareFunc - func type for consumer middleware
type MiddlewareFunc func(next MessageHandleFunc) MessageHandleFunc

// MessageConsumer - interface for message consumer
type MessageConsumer interface {
	Start(ctx context.Context) error
	Pause(ctx context.Context) error
	Resume(ctx context.Context) error

	AddHandler(ctx context.Context, topic string, h MessageHandleFunc) error
	Use(h MiddlewareFunc)
	Ready() bool
	Init(ctx context.Context) error
	Close(ctx context.Context) error
}

type db interface {
	Execute(ctx context.Context, statement string, args ...interface{}) error
	GetNextID(ctx context.Context, statement string) (int64, error)
}

type consumer struct {
	infrastructure.SugarLogger
	keepRunning      bool
	brokers          []string
	groupName        string
	config           *sarama.Config
	topics           []string
	handlers         map[string]MessageHandleFunc
	middleware       []MiddlewareFunc
	consumptionState bool
	mCh              chan bool
	cg               sarama.ConsumerGroup
	db               db
}

func NewConsumer(ctx context.Context, serviceName string, kafkaConfig KafkaConfig, db db) (MessageConsumer, error) {
	var target consumer
	if kafkaConfig.LogSarama {
		sarama.Logger = infrastructure.GetSaramaLogger(ctx)
	}
	target.brokers = kafkaConfig.BrokerList
	target.db = db
	target.groupName = serviceName

	if err := target.Init(ctx); err != nil {
		return nil, err
	}
	return &target, nil
}

func (s *consumer) Init(ctx context.Context) error {
	var err error
	s.keepRunning = true
	s.consumptionState = consumptionStopped
	s.handlers = make(map[string]MessageHandleFunc)
	s.mCh = make(chan bool)

	s.config = sarama.NewConfig()

	s.config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	s.config.Consumer.Offsets.Initial = sarama.OffsetNewest

	s.Use(prepareLoggerMiddleware)
	s.Use(logIncomingMessageMiddleware)

	s.cg, err = sarama.NewConsumerGroup(s.brokers, s.groupName, s.config)
	if err != nil {
		s.LogError(ctx, "Error creating consumer group client", err)
		return err
	}
	s.LogDebug(ctx, "consumer group created")
	return nil
}

func (s *consumer) Start(ctx context.Context) error {
	handler := consumerHandler{
		ready:      make(chan bool),
		handlers:   s.handlers,
		middleware: s.middleware,
		db:         s.db,
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			s.LogDebug(ctx, "try to join to consumer group")
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims

			if err := s.cg.Consume(ctx, s.topics, &handler); err != nil {
				s.LogError(ctx, "can't join to consumer group", err)
				time.Sleep(time.Second * 5)
			}
			// check if context was canceled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			handler.ready = make(chan bool)
		}
	}()

	<-handler.ready
	s.consumptionState = consumptionStarted
	s.LogInfo(ctx, "Sarama consumer up and running")

	for s.keepRunning {
		select {
		case <-ctx.Done():
			s.LogInfo(ctx, "Sarama consumer terminating: context canceled")
			s.keepRunning = false
		case <-s.mCh:
			s.toggleConsumptionFlow(ctx, s.cg)
		}
	}
	wg.Wait()
	_ = s.Close(ctx)
	return nil
}

func (s *consumer) Ready() bool {
	return s.consumptionState
}

func (s *consumer) AddHandler(ctx context.Context, topic string, h MessageHandleFunc) error {
	if topic == "" {
		s.LogError(ctx, "topic name is empty", ErrBadParam)
		return ErrBadParam
	}
	if h == nil {
		s.LogError(ctx, "can't find any handler ", ErrBadParam)
		return ErrBadParam
	}
	s.topics = append(s.topics, topic)
	s.handlers[topic] = h
	return nil
}

func (s *consumer) Pause(_ context.Context) error {
	s.mCh <- true
	return nil
}

func (s *consumer) Resume(_ context.Context) error {
	s.mCh <- true
	return nil
}

func (s *consumer) Use(h MiddlewareFunc) {
	s.middleware = append(s.middleware, h)
}

func (s *consumer) Close(ctx context.Context) error {
	if err := s.cg.Close(); err != nil {
		s.LogPanic(ctx, "Error closing consumer group", err)
	}
	return nil
}

func (s *consumer) toggleConsumptionFlow(ctx context.Context, cg sarama.ConsumerGroup) {
	if s.consumptionState {
		cg.ResumeAll()
		s.LogInfo(ctx, "Resuming consumption")
	} else {
		cg.PauseAll()
		s.LogInfo(ctx, "Pausing consumption")
	}
	s.consumptionState = !s.consumptionState
}

//-----------------------------------------------------------//

type consumerHandler struct {
	infrastructure.SugarLogger
	handlers   map[string]MessageHandleFunc
	middleware []MiddlewareFunc
	ready      chan bool
	db         db
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (h *consumerHandler) Setup(_ sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(h.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	log := infrastructure.GetBaseLogger(session.Context())
	log.Info().Msg(fmt.Sprintf("Consumer claim started(topic, partition,initial offset): %s, %d,%d", claim.Topic(), claim.Partition(), claim.InitialOffset()))
	for {
		select {
		case message := <-claim.Messages():
			messageHandler := h.applyMiddleware(h.handlers[message.Topic])
			//
			err := func() error {
				// pass new context to the handler!
				err := messageHandler(context.Background(), *message)
				if err != nil {
					log.Error().Err(err).Msg("error while message processing")
					// errorHandler must be started with new context!
					_, err2 := h.errorHandler(context.Background(), message, err)
					if err2 != nil {
						log.Error().Err(err2).Msg("db error")
						return err2
					}
					return nil
				}
				return nil
			}()
			if err != nil {
				log.Error().Err(err).Msg("can't process message")
			} else {
				session.MarkMessage(message, "")
			}

		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/Shopify/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		}
	}
}

func (h *consumerHandler) errorHandler(ctx context.Context, message *sarama.ConsumerMessage, occurredErr error) (int64, error) {
	id, err := h.db.GetNextID(ctx, sequenceNextIDInErrorMessage)
	if err != nil {
		return 0, err
	}
	//

	err = h.db.Execute(ctx, addKafkaInErrorMessage,
		id,

		message.Headers,
		h.headersToJSON(ctx, message.Headers),
		message.Timestamp,
		message.BlockTimestamp,
		message.Key,
		message.Value,
		message.Topic,
		message.Partition,
		message.Offset,
		occurredErr.Error(),
		time.Now())
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (h *consumerHandler) applyMiddleware(hf MessageHandleFunc) MessageHandleFunc {
	for i := len(h.middleware) - 1; i >= 0; i-- {
		hf = h.middleware[i](hf)
	}
	return hf
}

func (h *consumerHandler) headersToJSON(ctx context.Context, obj []*sarama.RecordHeader) []byte {
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
