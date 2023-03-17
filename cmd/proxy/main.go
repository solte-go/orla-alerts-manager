package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"rabbitmq/lab-soltegm.com/src/api"
	"rabbitmq/lab-soltegm.com/src/api/handlers/metrics/proxymetrics"
	schedulerHandler "rabbitmq/lab-soltegm.com/src/api/handlers/scheduler"
	"rabbitmq/lab-soltegm.com/src/config"
	"rabbitmq/lab-soltegm.com/src/logging"
	v1 "rabbitmq/lab-soltegm.com/src/queue/rabbitmq/v1"
	"rabbitmq/lab-soltegm.com/src/scheduler"
	"rabbitmq/lab-soltegm.com/src/toolbox/db"
	"syscall"
	"time"

	"go.uber.org/zap"

	_ "rabbitmq/lab-soltegm.com/src/scheduler/tasks/testtask"
)

var (
	env string
)

func init() {
	flag.StringVar(&env, "env", "dev", `Set's run environment. Possible values are "dev" and "prod"`)
	flag.Parse()
}

func waitQuitSignal(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		quit := make(chan os.Signal, 1)
		defer close(quit)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		<-quit
		cancel()
	}()

	return ctx
}

func main() {

	ctx := waitQuitSignal(context.Background())

	conf, err := config.LoadConf(env)
	if err != nil {
		panic(fmt.Sprintf("Error load config: %s", err.Error()))
	}

	logger, err := logging.NewLogger(conf.Logging)
	if err != nil {
		panic(fmt.Sprintf("Can't initialize logger: %s", err.Error()))
	}

	logger.Log(zap.DebugLevel, "Starting Proxy")

	defer logger.Sync() //nolint
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	err = v1.InitiateConfiguration(
		[]*config.RabbitMQ{
			conf.RabbitQueues.MainQueue,
		},
		logger, false,
	)
	if err != nil {
		logger.Fatal("Can't initialize RabbitMQ", zap.Error(err))
	}

	publisher, err := v1.GetConnection(conf.RabbitQueues.MainQueue.ConnName)
	if err != nil {
		logger.Fatal("Can't get RabbitMQ connection", zap.Error(err))
	}

	newStore, err := db.InitDatabase(ctx, conf, logger)
	if err != nil {
		logger.Fatal("Can't initialize store", zap.Error(err))
	}

	proxyMetrics := proxymetrics.NewSourcesProxy()

	sch := scheduler.New()

	server := api.NewServer(logger)
	go server.Run(
		ctx,
		conf.Server.Port,
		proxyMetrics,
		&schedulerHandler.SchedulerHandler{Scheduler: &sch},
	)

	for _, taskDesc := range conf.Scheduler.Tasks {
		logger.Info(taskDesc[0])
		durationString := taskDesc[2]
		duration, err := time.ParseDuration(durationString)
		if err != nil {
			logger.Fatal("Cant parse duration", zap.Error(err))
			return
		}

		startTime, err := sch.ParseStartTime(taskDesc[1])
		if err != nil {
			logger.Fatal("Cant parse start time", zap.Error(err))
			return
		}

		taskName := taskDesc[0]
		constructor, exists := scheduler.Registry[taskName]
		if !exists {
			logger.Fatal("Task not registered", zap.String("task", taskName))
			return
		}

		var task scheduler.Runnable
		task, err = constructor(duration, newStore, conf.Tasks)
		if err != nil {
			logger.Fatal("Cant create task", zap.Error(err))
			return
		}

		sch.AddScheduled(taskName, startTime, duration, task)
	}

    sch.Run(ctx, 2*time.Second, publisher)
}
