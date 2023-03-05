package worker

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

func (w *Worker) databaseWriteAlerts(ctx context.Context) error {

	var (
		delayJitter   = (rand.Float64() - 0.5) / 5
		delayDuration = w.job.connRetryDelay
		err           error
	)

	for {
		reqDelayJitter := time.Duration((1 + delayJitter) * float64(delayDuration))
		err = w.db.SharedDB.NewAlert(ctx, w.job.alerts)
		if err == nil {
			return nil
		}

		w.logger.Error(fmt.Sprintf("databaseWriteAlerts() call failed, retrying in %v", w.job.connRetryDelay), zap.Error(err))
		//TODO error metrics

		select {
		case <-time.After(reqDelayJitter):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
