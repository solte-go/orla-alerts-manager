package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"orla-alert/solte.lab/src/model"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// AckProcessedMessages will send acknowledged for processed messages to server
func (qc *QueueConnector) AckProcessedMessages() error {
	err := qc.Consumer.delivery.Ack(true)
	if err != nil {
		return err
	}

	return nil
}

// RequeueNotAcknowledgedMessages Will try to re-queue all not acknowledged messages. Will return err in delivery channel is nil
func (qc *QueueConnector) RequeueNotAcknowledgedMessages() error {
	deadLine, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var (
		delayJitter   = (rand.Float64() - 0.5) / 5
		delayDuration = 2 * time.Second
		err           error
	)

	for {
		reqDelayJitter := time.Duration((1 + delayJitter) * float64(delayDuration))
		if qc.Consumer.delivery != nil {
			err = qc.Consumer.delivery.Nack(true, true)
			if err != nil {
				return err
			} else {
				break
			}
		}

		qc.logger.Info(ErrQueueIsNotReady.Error())

		select {
		case <-time.After(reqDelayJitter):
		case <-deadLine.Done():
			return ErrQueueIsNotConfigured
		}
	}
	return nil
}

func (qc *QueueConnector) StartConsumer(workerPipeline chan model.Alert, errorChan chan error) error {
	if qc.CheckConnection() {
		err := qc.connect()
		if err != nil {
			return err
		}
	}

	qc.Consumer.workerPipeline = workerPipeline
	qc.Connection.errorChan = errorChan

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
				qc.Connection.errorChan <- err
				return
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
