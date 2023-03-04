package db

import (
	"context"
	"rabbitmq/lab-soltegm.com/src/logging"
	"rabbitmq/lab-soltegm.com/src/model"
	"rabbitmq/lab-soltegm.com/src/toolbox/db/contract"
	"testing"
	"time"

	"rabbitmq/lab-soltegm.com/src/config"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type storeTests struct {
	db     *DB
	logger *zap.Logger
	suite.Suite
}

func (suite *storeTests) Test_NewAlerts() {
	timeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer cancel()

	var alerts []*model.Alert
	newAlert := model.GetTestAlert()
	alerts = append(alerts, &newAlert)

	err := suite.db.SharedDB.NewAlert(timeout, alerts)
	suite.Suite.Error(err)

	results, err := suite.db.SharedDB.GetAllAlerts(timeout)
	suite.Suite.Error(err)

	suite.Suite.Equal(newAlert.HashID, results[0].HashID)
}

func (suite *storeTests) SetupSuite() {
	conf := config.InitConfig()
	conf.Logging.LogLevel = "info"
	logger, err := logging.NewLogger(conf.Logging)
	suite.Suite.NoError(err)

	newStore, err := InitDatabase(context.TODO(), conf, logger)
	suite.NoError(err)

	suite.db = newStore
	suite.logger = logger
}

func (suite *storeTests) TearDownTest() {
	err := suite.db.SharedDB.(contract.TeardownDB).ClearDatabase(context.TODO())
	suite.Suite.NoError(err)
}

func TestMongoDBSuite(t *testing.T) {
	suite.Run(t, new(storeTests))
}
