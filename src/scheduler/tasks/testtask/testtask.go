package testtask

import (
	"context"
	"orla-alerts/solte.lab/src/config"
	"orla-alerts/solte.lab/src/model"
	v1 "orla-alerts/solte.lab/src/queue/rabbitmq/v1"
	"orla-alerts/solte.lab/src/scheduler"
	"orla-alerts/solte.lab/src/toolbox/db"
	"time"

	"go.uber.org/zap"
)

const taskName = "test_task"

func init() {
	_ = scheduler.Add(taskName, newTestTask)
}

type Publisher interface {
	PrepareAndPublishMessage(ctx context.Context, routingKey string, entry model.Alert) error
	DefaultRouting() string
}

type testTask struct {
	logger    *zap.Logger
	publisher Publisher
}

func newTestTask(t time.Duration, db *db.DB, conf *config.Tasks) (scheduler.Runnable, error) {
	logger := zap.L().Named(taskName)

	rbInstance, err := v1.GetConnection(conf.ExampleTask.RabbitConnectionName)
	if err != nil {
		return nil, err
	}

	return &testTask{
		publisher: rbInstance,
		logger:    logger,
	}, nil
}

func (ts *testTask) Run() error {
	ts.logger.Debug("Starting task: test_task")

	time.Sleep(5 * time.Second)

	for i := 0; i < 1200; i++ {
		entry := model.GetTestAlert()
		err := ts.publisher.PrepareAndPublishMessage(context.TODO(), ts.publisher.DefaultRouting(), entry)
		if err != nil {
			ts.logger.Error("error", zap.Error(err))
		}
		time.Sleep(10 * time.Millisecond)
	}

	ts.logger.Debug("Finished task: test_task")
	return nil
}
