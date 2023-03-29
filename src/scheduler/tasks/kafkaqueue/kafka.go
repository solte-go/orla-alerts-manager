package kafka

import (
	"context"
	"orla-alert/solte.lab/src/config"
	"orla-alert/solte.lab/src/model"
	v1 "orla-alert/solte.lab/src/queue/kafkaqueue/v1"
	"orla-alert/solte.lab/src/scheduler"
	"orla-alert/solte.lab/src/toolbox/db"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
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
	var msgs = make([]*kafka.Message, 100)
	//time.Sleep(5 * time.Second)

	for i := 0; i < 1200; i++ {
		entry := model.GetTestAlert()

		msg := &kafka.Message{
			Key:   []byte(entry.Info.Name),
			Value: entry.ToJSON(),
		}
		msgs = append(msgs, msg)

		err := kt.kafkaPublisher.SendMessages(context.TODO(), msgs)
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
