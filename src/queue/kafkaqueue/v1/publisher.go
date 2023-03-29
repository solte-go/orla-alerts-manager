package kafkaQueueV1

import (
	"context"
	"orla-alert/solte.lab/src/config"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"time"

	"go.uber.org/zap"
)

var connectionPool = make(map[string]*Publisher)

type Publisher struct {
	conf     *config.Kafka
	producer *kafka.Producer
	logger   *zap.Logger
}

type sendResult struct {
	err error
	idx int
}

type messageWrapper struct {
	msg *kafka.Message
	idx int
}

func GetKafkaPublisher(connName string) (*Publisher, error) {
	if publisher, ok := connectionPool[connName]; ok {
		return publisher, nil
	}

	return nil, ErrNoConnectionWithProvidedName
}

func NewKafkaPublisher(conf *config.Kafka) (*Publisher, error) {
	var (
		producer        *kafka.Producer
		err             error
		publisherLogger = zap.L().Named("kafka_publisher")
		delay           = 1 * time.Second
		attempt         = 0
	)

	if conf.ConnectionName == "" {
		return nil, ErrEmptyConnectionName
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			attempt++
			delay *= 2
			publisherLogger.Info("Try connect to v1", zap.Int("attempt", attempt))

			producer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": conf.Brokers})
			if err == nil {

				p := &Publisher{
					conf:     conf,
					producer: producer,
					logger:   publisherLogger,
				}
				connectionPool[conf.ConnectionName] = p
				return p, err
			}
		}
	}
}

func (p *Publisher) sendMessage(ctx context.Context, message messageWrapper, resultChan chan sendResult) {
	deliveryChan := make(chan kafka.Event)
	defer close(deliveryChan)

	err := p.producer.Produce(message.msg, deliveryChan)
	if err != nil {
		resultChan <- sendResult{idx: message.idx, err: err}
		return
	}

	select {
	case <-ctx.Done():
		resultChan <- sendResult{idx: message.idx, err: ctx.Err()}
		<-deliveryChan
		return

	case event := <-deliveryChan:
		msg := event.(*kafka.Message)
		if msg.TopicPartition.Error != nil {
			resultChan <- sendResult{idx: message.idx, err: msg.TopicPartition.Error}
			return
		}

		resultChan <- sendResult{idx: message.idx, err: nil}
	}
}

func (p *Publisher) SendMessages(ctx context.Context, messages []*kafka.Message) []error {
	resultChan := make(chan sendResult, len(messages))
	defer close(resultChan)

	for index, message := range messages {
		go p.sendMessage(ctx, messageWrapper{msg: message, idx: index}, resultChan)
	}

	result := make([]error, len(messages))
	for range messages {
		receiver := <-resultChan
		result[receiver.idx] = receiver.err
	}

	return result
}
