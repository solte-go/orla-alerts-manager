package kafka

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
	"rabbitmq/lab-soltegm.com/src/config"
	"rabbitmq/lab-soltegm.com/src/model"
	"rabbitmq/lab-soltegm.com/src/queue/v1"
	"rabbitmq/lab-soltegm.com/src/scheduler"
	"rabbitmq/lab-soltegm.com/src/toolbox/db"
	"time"
)

const taskName = "kafka_test_task"

func init() {
	_ = scheduler.Add(taskName, newKafkaTestTask)
}

type kafkaTestTask struct {
	logger         *zap.Logger
	kafkaPublisher *v1.Publisher
}

func newKafkaTestTask(t time.Duration, db *db.DB, conf *config.Tasks) (scheduler.Runnable, error) {
	logger := zap.L().Named(taskName)

	publisher, err := v1.GetKafkaPublisher(conf.KafkaConnectionName)
	if err != nil {
		return nil, err
	}

	return &kafkaTestTask{
		kafkaPublisher: publisher,
		logger:         logger,
	}, nil
}

func (kt *kafkaTestTask) Run() error {
	kt.logger.Debug("Starting task: test_task")

	time.Sleep(5 * time.Second)

	for i := 0; i < 1200; i++ {
		entry := model.GetTestAlert()

		msg := &kafka.Message{
			Key:   []byte(entry.Info.Name),
			Value: entry.ToJSON(),
		}

		err := kt.kafkaPublisher.SendMessages(context.TODO(), []*kafka.Message{msg})
		for _, e := range err {
			if e != nil {
				kt.logger.Error("error", zap.Error(e))
			}

		}
		time.Sleep(10 * time.Millisecond)
	}
	kt.logger.Debug("Finished task: test_task")
	return nil
}
