package worker

import (
	"context"
	"errors"

	"time"

	"rabbitmq/lab-soltegm.com/src/config"
	"rabbitmq/lab-soltegm.com/src/model"
	v1 "rabbitmq/lab-soltegm.com/src/queue/rabbitmq/v1"
	"rabbitmq/lab-soltegm.com/src/toolbox/db"

	"go.uber.org/zap"
)

type jobChain func(ctx context.Context, id uint64) error

func runJobChainFunc(ctx context.Context, id uint64, fns ...jobChain) error {
	for _, fn := range fns {
		if err := fn(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

type Worker struct {
	logger   *zap.Logger
	db       *db.DB
	consumer *v1.QueueConnector
	Interval time.Duration
	job      job
}

func NewWorker(conf config.Config, logger *zap.Logger, conn *v1.QueueConnector, db *db.DB) *Worker {
	return &Worker{
		logger:   logger,
		consumer: conn,
		db:       db,
		Interval: conf.Worker.TickerInterval,
		job: job{
			status:           notRunning,
			bunchSize:        conf.Worker.BunchSize,
			timeout:          conf.Worker.TaskTimeout,
			currentBunchSize: 0,
			connRetryDelay:   conf.Worker.TaskConnDelay,
		},
	}
}

func (w *Worker) Run(ctx context.Context) error {
	deliveryChannel := make(chan model.Alert)
	errorChan := make(chan error)
	w.job.deliveryChan = deliveryChannel

	err := w.consumer.StartConsumer(deliveryChannel, errorChan)
	if err != nil {
		return err
	}

	err = w.consumer.RequeueNotAcknowledgedMessages()
	if err != nil {
		w.logger.Error("can't re-queue unacknowledged messages", zap.Error(err))
	}

	ticker := time.NewTicker(w.Interval)
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Exiting worker")
			ticker.Stop()
			return ctx.Err()

		case errChan := <-errorChan:
			w.logger.Error("error", zap.Error(errChan))
			w.logger.Info("Exiting worker")
			return errChan

		case <-ticker.C:
			//TODO reset massages if previus run completed with error
			if w.job.ready() {
				go w.runJobInstance()
			}
		}
	}
}

func (w *Worker) runJobInstance() error {
	w.logger.Debug("Job in Progress")
	w.job.running()
	ctxWithTimeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(w.job.timeout))
	defer cancel()

	var data model.Alert

	for {
		select {
		case data = <-w.job.deliveryChan:
			//TODO add metrics
		case <-ctxWithTimeout.Done():
			w.job.returnedWithError()
			return errors.New("the task has reached the time limit, canceling")
		}

		w.job.addToBunch(data)

		if w.job.currentBunchSize >= w.job.bunchSize {

			//TODO chain of process will be here

			if err := w.databaseWriteAlerts(ctxWithTimeout); err != nil {
				w.job.returnedWithError()
				w.logger.Error("Error", zap.Error(err))
				return err
			}

			err := w.consumer.AckProcessedMessages()
			if err != nil {
				w.logger.Error("Error", zap.Error(err))
				err := w.consumer.ReconnectConsumer()
				if err != nil {
					w.logger.Error("Reconnection failed", zap.Error(err))
					w.job.returnedWithError()
					return err
				}
			}

			//resetting data structs for next cycle
			w.job.resetBunch()
			break
		}
	}
	w.job.completedSuccessfully()
	w.logger.Debug("Job Completed")
	return nil
}
