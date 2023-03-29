package v1

import (
	"go.uber.org/zap"
	"orla-alerts/solte.lab/src/config"
	"orla-alerts/solte.lab/src/logging"
	"orla-alerts/solte.lab/src/model"
	"time"

	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// TODO Check it tomorrow
type queueTests struct {
	suite.Suite
	rabbit *QueueConnector
	logger *zap.Logger
	conf   *config.Config
}

func (suite *queueTests) publishTestMessages(conn *QueueConnector) error {
	for i := 0; i < 2; i++ {
		entry := model.GetTestAlert()

		err := conn.PrepareAndPublishMessage(context.TODO(), conn.DefaultRouting(), entry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (suite *queueTests) Test_PublisherReconnection() {
	testPublisher, err := GetConnection(suite.conf.RabbitQueues.MainQueue.ConnName)
	suite.Suite.NoError(err)
	errCh := make(chan error)

	err = testPublisher.StartPublisher(errCh)
	suite.Suite.NoError(err)

	err = suite.publishTestMessages(testPublisher)
	suite.Suite.NoError(err)

	testPublisher.CloseConnection()

	err = suite.publishTestMessages(testPublisher)
	suite.Suite.NoError(err)
}

func (suite *queueTests) Test_ConsumerReconnection() {
	var data []model.Alert
	errCh := make(chan error)

	testPublisher, err := GetConnection(suite.conf.RabbitQueues.MainQueue.ConnName)
	suite.Suite.NoError(err)

	err = testPublisher.StartPublisher(errCh)
	suite.Suite.NoError(err)

	err = suite.publishTestMessages(testPublisher)
	suite.Suite.NoError(err)

	ch := make(chan model.Alert)
	chErr := make(chan error)

	testConsumer, err := GetConnection("test_consumer")
	suite.Suite.NoError(err)

	err = testConsumer.StartConsumer(ch, chErr)
	suite.Suite.NoError(err)

	msg := <-ch
	data = append(data, msg)
	suite.logger.Debug("data received", zap.String(msg.Info.Brand, msg.Event.Severity))
	suite.Equal("Orla", msg.Info.Brand)

	err = testConsumer.AckProcessedMessages()
	suite.Suite.NoError(err)

	testPublisher.CloseConnection()

	time.Sleep(1 * time.Second)

	err = testConsumer.RequeueNotAcknowledgedMessages()
	suite.Suite.NoError(err)

	msg = <-ch
	data = append(data, msg)
	suite.logger.Debug("data received", zap.String(msg.Info.Brand, msg.Event.Severity))
	suite.Equal(2, len(data))
}

func (suite *queueTests) SetupSuite() {
	conf := config.InitConfig()
	suite.conf = &conf

	logger, err := logging.NewLogger(conf.Logging)
	if err != nil {
		panic(fmt.Sprintf("Can't initialize logger: %s", err.Error()))
	}
	suite.logger = logger

	err = InitiateConfiguration(
		[]*config.RabbitMQ{conf.RabbitQueues.MainQueue}, logger, false)
	suite.Suite.NoError(err)

	err = InitiateConfiguration(
		[]*config.RabbitMQ{conf.RabbitQueues.DelayedQueue}, logger, true)
	suite.Suite.NoError(err)

	_, err = NewConnection(
		"test_consumer",
		[]string{conf.RabbitQueues.MainQueue.Queue, conf.RabbitQueues.DelayedQueue.Queue},
		conf.RabbitQueues.MainQueue,
		logger,
	)

	if err != nil {
		logger.Fatal("Can't connect to RabbitMQ", zap.Error(err))
	}
}

func (suite *queueTests) TearDownSuite() {
	conn, err := GetConnection(suite.conf.RabbitQueues.MainQueue.ConnName)
	suite.Suite.NoError(err)

	conn.Connection.channel.QueueDelete(suite.conf.RabbitQueues.MainQueue.Queue, false, false, true)
	conn.Connection.channel.QueueDelete(suite.conf.RabbitQueues.DelayedQueue.Queue, false, false, true)
}

func TestPublisher(t *testing.T) {
	suite.Run(t, &queueTests{})
}
