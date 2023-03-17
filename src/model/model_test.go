package model

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

const alerts = 10000

type alertsTest struct {
	suite.Suite
}

func (suite *alertsTest) Test_Alerts() {
	var al Alert
	testMap := make(map[uint64]Alert, alerts)

	for i := 0; i < alerts; i++ {
		al = GetTestAlert()
		if _, exist := testMap[al.Info.ID]; exist {
			suite.Fail("id shouldn't repiete itself")
		} else {
			testMap[al.Info.ID] = al
		}
	}

	suite.NotNil(testMap, "Should not be empty")
}

func TestAlertsSuite(t *testing.T) {
	suite.Run(t, new(alertsTest))
}
