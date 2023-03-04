package v1

import (
	"rabbitmq/lab-soltegm.com/src/config"
	"rabbitmq/lab-soltegm.com/src/logging"
	"rabbitmq/lab-soltegm.com/src/model"

	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// TODO Check it tomorrow
type queueTests struct {
	suite.Suite
	rabbit *QueueConnector
	conf   *config.RabbitMQ
}

func (suite *queueTests) Test_QueueFunctional() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := suite.rabbit.StartPublisher()
	suite.Suite.NoError(err)

	alert := model.GetTestAlert()

	err = suite.rabbit.PrepareAndPublishMessage(ctx, Default, alert)
	suite.Suite.NoError(err)

	//ch := make(chan model.NRDQueue)
	//suite.rabbit.workerPipeline = ch
	//suite.rabbit.StartConsumer(ch)
	//
	//data := <-ch
	//fmt.Println(data)

	//_, err = suite.rabbit.Connection.channel.QueuePurge(suite.conf.Queue, false)
	//suite.Suite.NoError(err)
}

func (suite *queueTests) Test_ConsumerReconnection() {
	ch := make(chan model.Alert)
	err := suite.rabbit.StartConsumer(ch)
	suite.Suite.NoError(err)

	err = suite.rabbit.ReconnectConsumer()
	suite.Suite.NoError(err)

	newConnection := suite.rabbit.Connection.client.ConnectionState()
	suite.Equal(false, newConnection.DidResume)
}

func (suite *queueTests) Test_PublisherReconnection() {
	err := suite.rabbit.StartPublisher()
	suite.Suite.NoError(err)

	err = suite.rabbit.ReconnectPublisher()
	suite.Suite.NoError(err)

	newConnection := suite.rabbit.Connection.client.ConnectionState()
	suite.Equal(false, newConnection.DidResume)
}

func (suite *queueTests) SetupTest() {
	conf := config.InitConfig()

	logger, err := logging.NewLogger(conf.Logging)
	if err != nil {
		panic(fmt.Sprintf("Can't initialize logger: %s", err.Error()))
	}

	err = InitiateConfiguration(
		[]*config.RabbitMQ{
			conf.RabbitQueues.MainQueue,
		},
		logger, false,
	)
	suite.Suite.NoError(err)

	conn, err := GetConnection(conf.RabbitQueues.MainQueue.ConnName)
	suite.Suite.NoError(err)

	suite.rabbit = conn
	suite.conf = conf.RabbitQueues.MainQueue
}

//func (suite *queueTests) TearDownTest() {
//	//_, err := suite.rabbit.Connection.channel.QueuePurge(suite.conf.Queue, false)
//	//suite.Suite.NoError(err)
//}

func (suite *queueTests) TearDownSuite() {
	_, err := suite.rabbit.Connection.channel.QueuePurge(suite.conf.Queue, false)
	suite.Suite.NoError(err)

	suite.rabbit.Shutdown()
}

func TestPublisher(t *testing.T) {
	suite.Run(t, &queueTests{})
}
