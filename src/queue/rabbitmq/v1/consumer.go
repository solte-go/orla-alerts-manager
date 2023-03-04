package v1

import (
	"encoding/json"
	"fmt"
	"rabbitmq/lab-soltegm.com/src/model"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// GetDeliveryChannel Return delivery channel to be able ACK delivery of message outside of RabbitMQ package
func (qc *QueueConnector) GetDeliveryChannel() *amqp.Delivery {
	return qc.Consumer.delivery
}

func (qc *QueueConnector) StartConsumer(workerPipeline chan model.Alert) error {
	if qc.CheckConnection() {
		err := qc.connect()
		if err != nil {
			return err
		}
	}

	qc.Consumer.workerPipeline = workerPipeline

	if err := qc.Connection.channel.Qos(
		qc.conf.Qos, // prefetch count
		0,           // psrefetch size
		true,        // global
	); err != nil {
		qc.logger.Error("Failed to set QoS", zap.Error(err))
		return err
	}

	deliveries, err := qc.ConsumerMessages()
	if err != nil {
		qc.logger.Error("can't restart Consumer()", zap.Error(err))
		return err
	}
	for queue, delivery := range deliveries {
		go qc.HandleConsumedDeliveries(queue, delivery, workerPipeline, qc.messageHandler)
	}
	return nil
}

// HandleConsumedDeliveries handles the consumed deliveries from the queues. Should be called only for a consumer connection
func (qc *QueueConnector) HandleConsumedDeliveries(
	queue string,
	delivery <-chan amqp.Delivery,
	wp chan model.Alert,
	fn func(QueueConnector, string, <-chan amqp.Delivery, chan model.Alert),
) {
	for {
		go fn(*qc, queue, delivery, wp)
		if err := <-qc.Connection.err; err != nil {
			qc.ReconnectConsumer()
			deliveries, err := qc.ConsumerMessages()
			if err != nil {
				qc.logger.Error("can't restart Consumer()", zap.Error(err))
			}
			delivery = deliveries[queue]
		}
	}
}

func (qc *QueueConnector) ConsumerMessages() (map[string]<-chan amqp.Delivery, error) {
	m := make(map[string]<-chan amqp.Delivery)
	for _, queue := range qc.Connection.queues {
		deliveries, err := qc.Connection.channel.Consume(
			queue,
			"",
			qc.conf.AutoAck,
			false,
			false,
			false,
			nil)
		if err != nil {
			return nil, err
		}
		m[queue] = deliveries
	}
	return m, nil
}

func (qc *QueueConnector) messageHandler(conn QueueConnector, queue string, deliveries <-chan amqp.Delivery, wp chan model.Alert) {
	var data model.Alert
	for d := range deliveries {
		//need to be able ACK messages outside of package
		//TODO conf for common setting in rabbit MQ
		qc.Consumer.delivery = &d
		if qc.conf.Verbose == true {
			qc.logger.Log(zap.InfoLevel,
				fmt.Sprintf("got %dB delivery: [%v] %q",
					len(d.Body),
					d.DeliveryTag,
					d.Body),
			)
			if err := json.Unmarshal(d.Body, &data); err != nil {
				qc.logger.Error("error during Unmarshal process", zap.Error(err))
			}
			wp <- data
		} else {
			if err := json.Unmarshal(d.Body, &data); err != nil {
				qc.logger.Error("error during Unmarshal process", zap.Error(err))
			}
			wp <- data
		}
	}
}
