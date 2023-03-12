package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"rabbitmq/lab-soltegm.com/src/model"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (qc *QueueConnector) StartPublisher(errCh chan error) error {
	if qc.CheckConnection() {
		err := qc.connect()
		if err != nil {
			return err
		}
	}

	qc.Connection.errorChan = errCh
	qc.Publisher.confirms = qc.Connection.channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	qc.Publisher.publishes = make(chan uint64, 8)
	qc.Publisher.done = make(chan struct{})

	qc.Connection.reliable = true

	if qc.Connection.reliable {
		if err := qc.Connection.channel.Confirm(false); err != nil {
			return fmt.Errorf("channel could not be put into confirm mode: %s", err)
		}
		go qc.confirmHandler(qc.Publisher.done, qc.Publisher.publishes, qc.Publisher.confirms)
	}
	return nil
}

func (qc *QueueConnector) confirmHandler(done chan struct{}, publishes chan uint64, confirms chan amqp.Confirmation) {
	m := make(map[uint64]bool)
	for {
		select {
		case <-done:
			qc.logger.Debug("confirmHandler is stopping")
			return
		case publishSeqNo := <-publishes:
			//qc.logger.Debug(fmt.Sprintf("waiting for confirmation of %d", publishSeqNo))
			m[publishSeqNo] = false
		case confirmed := <-confirms:
			if confirmed.DeliveryTag > 0 {
				if confirmed.Ack {
					//qc.logger.Debug(fmt.Sprintf("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag))
				} else {
					//qc.logger.Debug(fmt.Sprintf("failed delivery of delivery tag: %d", confirmed.DeliveryTag))
				}
				delete(m, confirmed.DeliveryTag)
			}
		}
		if len(m) > 1 {
			//qc.logger.Debug(fmt.Sprintf("outstanding confirmations: %d", len(m)))
		}
	}
}

func (qc *QueueConnector) PrepareAndPublishMessage(ctx context.Context, routingKey string, entry model.Alert) error {
	var header amqp.Table

	messageBody, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	message := amqp.Publishing{
		Headers:      amqp.Table{"type": "default"},
		ContentType:  "text/plain",
		DeliveryMode: amqp.Transient,
		Timestamp:    time.Now(),
		Body:         messageBody,
	}

	if routingKey == qc.DelayRouting() {
		header = amqp.Table{"x-delay": qc.conf.ExchangeArgs.Delay.Milliseconds()}
		message.Headers = header
	}

	err = qc.PublishMessage(ctx, routingKey, message)
	if err != nil {
		return err
	}

	return nil
}

// PublishMessage publishes a request to the amqp queue
func (qc *QueueConnector) PublishMessage(ctx context.Context, routingKey string, message amqp.Publishing) error {
	seqNo := qc.Connection.channel.GetNextPublishSeqNo()

	select { //non-blocking channel - if there is no error will go to default where we do nothing
	case err := <-qc.Connection.err:
		if err != nil {
			qc.logger.Warn("trying to reconnect Publisher")
			qc.ReconnectPublisher()
		}
	default:
	}

	//TODO need more tests
	if err := qc.Connection.channel.PublishWithContext(
		ctx,
		qc.Connection.exchange,
		routingKey,
		false,
		false,
		message,
	); err != nil {
		qc.ReconnectPublisher()
		return fmt.Errorf("error while publishing message: %s", err)
	}

	if qc.Connection.reliable {
		qc.Publisher.publishes <- seqNo
	}

	return nil
}
