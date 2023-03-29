package scheduler

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	v1 "orla-alerts/solte.lab/src/queue/rabbitmq/v1"
)

type Scheduler struct {
	jobSync syncContainer
	logger  *zap.Logger
}

func New() Scheduler {
	return Scheduler{
		jobSync: newSyncContainer(),
		logger:  zap.L().Named("scheduler"),
	}
}

func (s *Scheduler) AddScheduled(name string, firstRun time.Time, duration time.Duration, task Runnable) error {
	var err error
	s.jobSync.access(func(tasks map[string]*taskWrapper) {
		if _, exists := tasks[name]; exists {
			err = fmt.Errorf("task %q already registered", name)
			return
		}

		for {
			if time.Now().After(firstRun) {
				firstRun = firstRun.Add(duration)
			} else {
				break
			}
		}

		tasks[name] = &taskWrapper{
			Runnable: task,
			duration: duration,
			nextRun:  firstRun,
		}

	})

	return err
}

func (s *Scheduler) Run(ctx context.Context, resolution time.Duration, r *v1.QueueConnector) {

	ticker := time.NewTicker(resolution)

	errChan := make(chan error)
	err := r.StartPublisher(errChan)
	if err != nil {
		s.logger.Fatal("Can't start RabbitMQ Publisher", zap.Error(err))
	}

	for {
		s.jobSync.access(func(tasks map[string]*taskWrapper) {
			for n, t := range tasks {
				if t.Ready() {
					t.ScheduleNextRun()
					t.status = statusRunning
					t.lastRun = time.Now()
					go s.runTask(n, t)
				}
			}
		})

		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			continue
		}
	}
}

func (s *Scheduler) runTask(taskName string, t *taskWrapper) {
	defer func() {
		if r := recover(); r != nil {
			zap.L().Error("Recovered", zap.Any("message", r), zap.ByteString("stacktrace", debug.Stack()))
			s.jobSync.access(func(tasks map[string]*taskWrapper) {
				task := tasks[taskName]
				task.lastError = fmt.Errorf("recovered: %v", r)
				task.status = statusNotRunning
			})

		}
	}()

	zap.L().Info("", zap.String("Starting task:", taskName))
	err := t.Run()
	if err != nil {
		zap.L().Error("task exited with error", zap.String("task name", taskName), zap.Error(err))
	}
	zap.L().Info("", zap.String("Finished task:", taskName))

	s.jobSync.access(func(tasks map[string]*taskWrapper) {
		task := tasks[taskName]
		task.lastError = err
		task.status = statusNotRunning
	})
}

func (s *Scheduler) TasksListInfo() map[string]TaskInfo {
	info := map[string]TaskInfo{}

	s.jobSync.access(func(tasks map[string]*taskWrapper) {
		for name, task := range tasks {
			info[name] = task.TaskInfo()
		}
	})

	return info
}

func (s *Scheduler) TaskInfo(name string) (TaskInfo, error) {
	var (
		err  error
		info TaskInfo
	)

	s.jobSync.access(func(tasks map[string]*taskWrapper) {
		task, exists := tasks[name]
		if !exists {
			err = fmt.Errorf("task '%s' not registered", name)
			return
		}
		info = task.TaskInfo()
	})

	return info, err
}

func (s *Scheduler) RunOnNextTick(name string) error {
	var err error

	s.jobSync.access(func(tasks map[string]*taskWrapper) {
		task, exists := tasks[name]
		if !exists {
			err = fmt.Errorf("task '%s' not registered", name)
			return
		}

		if task.status == statusRunning {
			err = fmt.Errorf("task '%s' already running", name)
			return
		}

		tasks[name].runOnNextTick = true
	})

	return err
}

func (s *Scheduler) Cancel(name string) error {
	var err error

	s.jobSync.access(func(tasks map[string]*taskWrapper) {
		task, exists := tasks[name]
		if !exists {
			err = fmt.Errorf("task '%s' not registered", name)
			return
		}

		task.duration = 0
		task.nextRun = time.Time{}
	})

	return err
}

func (s *Scheduler) ScheduleTask(name string, timeToRun time.Time, interval time.Duration) error {
	var err error

	s.jobSync.access(func(tasks map[string]*taskWrapper) {
		task, exists := tasks[name]
		if !exists {
			err = fmt.Errorf("task '%s' not registered", name)
			return
		}

		for {
			if time.Now().After(timeToRun) {
				timeToRun = timeToRun.Add(interval)
			} else {
				break
			}
		}

		task.nextRun = timeToRun
		task.duration = interval
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) ParseStartTime(timeString string) (time.Time, error) {
	currentTime := time.Now()
	currentYear, currentMonth, currentDay := currentTime.Date()

	ts := strings.SplitN(timeString, ":", 2)
	hour, err := strconv.ParseInt(ts[0], 0, 0)
	if err != nil {
		return time.Time{}, err
	}
	minutes, err := strconv.ParseInt(ts[1], 0, 0)
	if err != nil {
		return time.Time{}, err
	}

	taskFirstRun := time.Date(currentYear, currentMonth, currentDay, int(hour), int(minutes), 0, 0, time.UTC)
	return taskFirstRun, nil
}
