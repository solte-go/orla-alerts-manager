package db

import (
	"context"
	"fmt"
	"orla-alert/solte.lab/src/logging"
	"orla-alert/solte.lab/src/model"
	"orla-alert/solte.lab/src/toolbox/db/contract"
	"testing"
	"time"

	"orla-alert/solte.lab/src/config"

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
	suite.Suite.NoError(err)

	results, err := suite.db.SharedDB.GetAllAlerts(timeout)
	suite.Suite.NoError(err)

	suite.Suite.Equal(fmt.Sprintf("%v", newAlert.Info.ID), fmt.Sprintf("%v", results[0].Info.ID))
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
