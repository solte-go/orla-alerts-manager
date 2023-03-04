package testtask

import (
	"context"
	"rabbitmq/lab-soltegm.com/src/config"
	"rabbitmq/lab-soltegm.com/src/model"
	v1 "rabbitmq/lab-soltegm.com/src/queue/rabbitmq/v1"
	"rabbitmq/lab-soltegm.com/src/scheduler"
	"rabbitmq/lab-soltegm.com/src/toolbox/db"
	"time"

	"go.uber.org/zap"
)

const taskName = "test_task"

func init() {
	_ = scheduler.Add(taskName, newTestTask)
}

type testTask struct {
	logger   *zap.Logger
	rabbitmq *v1.QueueConnector
}

func newTestTask(t time.Duration, db *db.DB, conf *config.Tasks) (scheduler.Runnable, error) {
	logger := zap.L().Named(taskName)

	connection, err := v1.GetConnection(conf.ExampleTask.RabbitConnectionName)
	if err != nil {
		return nil, err
	}

	return &testTask{
		rabbitmq: connection,
		logger:   logger,
	}, nil
}

func (ts *testTask) Run() error {
	ts.logger.Debug("Starting task: test_task")

	entry := model.GetTestAlert()

	time.Sleep(5 * time.Second)

	for i := 0; i < 100; i++ {
		err := ts.rabbitmq.PrepareAndPublishMessage(context.TODO(), v1.Default, entry)
		if err != nil {
			ts.logger.Error("error", zap.Error(err))
		}
		time.Sleep(10 * time.Millisecond)
	}

	ts.logger.Debug("Finished task: test_task")
	return nil
}
