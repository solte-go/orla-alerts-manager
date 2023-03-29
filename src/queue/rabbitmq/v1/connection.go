package v1

import (
	"context"
	"fmt"
	"math/rand"
	"orla-alert/solte.lab/src/config"
	"orla-alert/solte.lab/src/model"
	"time"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Header struct {
	HeaderType string
	Delay      int64
}

type connection struct {
	name      string
	client    *amqp.Connection
	channel   *amqp.Channel
	connChan  chan *amqp.Error
	errorChan chan error
	exchange  string
	queues    []string
	err       chan error
	reliable  bool
}

type consumer struct {
	workerPipeline chan model.Alert
	delivery       *amqp.Delivery
}

type publisher struct {
	publishes chan uint64
	confirms  chan amqp.Confirmation
	done      chan struct{}
}

// QueueConnector ...
type QueueConnector struct {
	Connection connection
	Consumer   consumer
	Publisher  publisher
	conf       *config.RabbitMQ
	logger     *zap.Logger
}

var (
	connectionPool = make(map[string]*QueueConnector)
)

// InitiateConfiguration will configure RabbitMQ exchanges & queues by connecting to RabbitMQ server.
// if removeAfterConfiguration is true, connection will be close and deleted from connection pool map after configuration.
// RabbitMQ will preserve configuration for future connections
func InitiateConfiguration(instances []*config.RabbitMQ, logger *zap.Logger, removeAfterConfiguration bool) error {
	for _, instance := range instances {
		conn, err := GetConnection(instance.ConnName)
		if err == ErrNoSuchConnection {
			conn, err = NewConnection(instance.ConnName, []string{instance.Queue}, instance, logger)
			if err != nil {
				return err
			}

			err = conn.declareExchange()
			if err != nil {
				return err
			}

			err = conn.bindQueue()
			if err != nil {
				return err
			}
		}

		if removeAfterConfiguration {
			if err := DisconnectAndRemoveConnection(instance.ConnName); err != nil {
				return err
			}

		}

	}
	return nil
}

// NewConnection Will create a new connection and add it to connection pool map
func NewConnection(name string, queues []string, conf *config.RabbitMQ, logger *zap.Logger) (*QueueConnector, error) {
	if c, ok := connectionPool[name]; ok {
		return c, nil
	}

	cfgConn := connection{
		name:     name,
		exchange: conf.Exchange,
		queues:   queues,
		err:      make(chan error),
		reliable: true,
	}

	c := &QueueConnector{
		Connection: cfgConn,
	}

	c.conf = conf
	c.logger = logger

	err := c.connect()
	if err != nil {
		return nil, err
	}

	connectionPool[name] = c
	return c, nil
}

// GetConnection retrieve connection from connection pool by provided name. If there is no connection will return err ErrNoSuchConnection
func GetConnection(name string) (*QueueConnector, error) {
	if name == "" {
		return nil, ErrEmptyConnectionString
	}

	if c, ok := connectionPool[name]; ok {
		return c, nil
	}
	return nil, ErrNoSuchConnection
}

// DisconnectAndRemoveConnection removes connection from connection pool
func DisconnectAndRemoveConnection(name string) error {
	if conn, ok := connectionPool[name]; ok {
		conn.Connection.client.Close()

		delete(connectionPool, name)
		return nil
	}
	return ErrNoSuchConnection
}

// connect internal function what creates a new connection to RabbitMQ
func (qc *QueueConnector) connect() error {
	var err error

	cfg := amqp.Config{Properties: amqp.NewConnectionProperties()}
	cfg.Properties.SetClientConnectionName(qc.Connection.name)

	qc.Connection.client, err = amqp.DialConfig(qc.conf.AmqpURI, cfg)
	if err != nil {
		return fmt.Errorf("can't creating rabbitmq connection %s", err.Error())
	}

	go func() {
		qc.Connection.connChan = qc.Connection.client.NotifyClose(make(chan *amqp.Error)) //Listen to NotifyClose
		<-qc.Connection.connChan
		qc.logger.Info(fmt.Sprintf("closing connection %s", qc.Connection.name))
		qc.Connection.err <- errors.New("connection closed")
	}()

	qc.Connection.channel, err = qc.Connection.client.Channel()
	if err != nil {
		return fmt.Errorf("channel: %s", err)
	}
	return nil
}

// declareExchange declaring Exchange if not declared
func (qc *QueueConnector) declareExchange() error {
	args := amqp.Table{}

	if qc.conf.ExchangeArgs.Key != "" {
		args[qc.conf.ExchangeArgs.Key] = qc.conf.ExchangeArgs.Value
	}

	if err := qc.Connection.channel.ExchangeDeclare(
		qc.conf.Exchange,     // name
		qc.conf.ExchangeType, // type
		qc.conf.Durable,      // durable
		false,                // auto-deleted
		false,                // internal
		false,                // noWait
		args,                 // arguments
	); err != nil {
		return fmt.Errorf("error in main exchange declare: %s", err)
	}
	return nil
}

// bindQueue declaring Queue if not Queue
func (qc *QueueConnector) bindQueue() error {
	for _, queue := range qc.Connection.queues {
		if _, err := qc.Connection.channel.QueueDeclare(
			qc.conf.Queue,
			qc.conf.Durable,
			false,
			false,
			false,
			nil); err != nil {
			return fmt.Errorf("error in declaring the queue %s", err)
		}
		if err := qc.Connection.channel.QueueBind(
			queue,
			qc.conf.BindingKey,
			qc.conf.Exchange,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("queue bind error: %s", err)
		}
	}
	return nil
}

// ReconnectConsumer reconnects & configured the connection for Consumer side of RabbitMQ
func (qc *QueueConnector) ReconnectConsumer() error {
	var (
		delayJitter   = (rand.Float64() - 0.5) / 5
		delayDuration = 3 * time.Second
	)

	qc.CloseConnection()

	deadLine, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		reqDelayJitter := time.Duration((1 + delayJitter) * float64(delayDuration))
		err := qc.connect()
		if err == nil {
			break
		}

		select {
		case <-time.After(reqDelayJitter):
		case <-deadLine.Done():
			qc.Connection.errorChan <- errFailedToStartConsumer
			return errFailedToStartConsumer
		}

		qc.logger.Error("an error when trying to recover Consumer connection", zap.Error(err))
		time.Sleep(reqDelayJitter)
	}

	err := qc.StartConsumer(qc.Consumer.workerPipeline, qc.Connection.errorChan)
	if err != nil {
		return err
	}

	return nil
}

// ReconnectPublisher reconnects & configured the connection for Publisher side of RabbitMQ
func (qc *QueueConnector) ReconnectPublisher() error {
	var (
		delayJitter   = (rand.Float64() - 0.5) / 5
		delayDuration = 3 * time.Second
		err           error
	)

	qc.CloseConnection()

	deadLine, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		reqDelayJitter := time.Duration((1 + delayJitter) * float64(delayDuration))
		err = qc.connect()
		if err == nil {
			break
		}
		qc.logger.Error("re-initiate Publisher connection error", zap.Error(err))

		select {
		case <-time.After(reqDelayJitter):
		case <-deadLine.Done():
			qc.Connection.errorChan <- errFailedToStartPublisher
			return errFailedToStartPublisher
		}
	}

	if err = qc.bindQueue(); err != nil {
		return err
	}

	if qc.Publisher.done != nil {
		qc.Publisher.done <- struct{}{}
	}

	err = qc.StartPublisher(qc.Connection.errorChan)
	if err != nil {
		return err
	}

	qc.logger.Info("Connection reestablished")

	return nil
}

func (qc *QueueConnector) CheckConnection() bool {
	return qc.Connection.client.IsClosed()
}

// Shutdown will delete connection from connections pull
func (qc *QueueConnector) Shutdown() error {
	if qc.Publisher.done != nil {
		close(qc.Publisher.done)
	}

	if err := DisconnectAndRemoveConnection(qc.Connection.name); err != nil {
		return err
	}

	return nil
}

func (qc *QueueConnector) CloseConnection() {

	err := qc.Connection.channel.Close()
	if err != nil {
		fmt.Println("channel", err)
		return
	}

	qc.Connection.client.Close()
}

// ClearQueue this function will clear all messages from provided queue
// mainly used for test purposes
func (qc *QueueConnector) ClearQueue() error {
	_, err := qc.Connection.channel.QueuePurge(qc.DelayRouting(), false)
	if err != nil {
		return err
	}
	return nil
}

func (qc *QueueConnector) DefaultRouting() string {
	return qc.conf.BindingKey
}

func (qc *QueueConnector) DelayRouting() string {
	return qc.conf.BindingKey
}
