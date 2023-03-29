package main

import (
	"context"
	"flag"
	"fmt"
	"orla-alert/solte.lab/src/config"
	"orla-alert/solte.lab/src/logging"
	v1 "orla-alert/solte.lab/src/queue/rabbitmq/v1"
	"orla-alert/solte.lab/src/toolbox/db"
	"orla-alert/solte.lab/src/worker"
	"os"
	"os/signal"

	"syscall"

	"go.uber.org/zap"
)

var (
	env string
)

func init() {
	flag.StringVar(&env, "env", "dev", `Set's run environment. Possible values are "dev" and "prod"`)
	flag.Parse()
}

func main() {
	//go func() {
	//	log.Fatal(http.ListenAndServe(":8099", nil))
	//}()

	ctx := waitQuitSignal(context.Background())

	conf, err := config.LoadConf(env)
	if err != nil {
		panic(fmt.Sprintf("error load config: %s", err.Error()))
	}

	logger, err := logging.NewLogger(conf.Logging)
	if err != nil {
		panic(fmt.Sprintf("Can't initialize logger: %s", err.Error()))
	}

	logger.Named("worker")

	defer logger.Sync()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	newDB, err := db.InitDatabase(ctx, conf, logger)
	if err != nil {
		logger.Fatal("Can't initialize store", zap.Error(err))
	}

	err = v1.InitiateConfiguration(
		[]*config.RabbitMQ{
			conf.RabbitQueues.MainQueue,
			conf.RabbitQueues.DelayedQueue,
		}, logger, true,
	)
	if err != nil {
		logger.Fatal("Can't initialize RabbitMQ", zap.Error(err))
	}

	rabbitConsumer, err := v1.NewConnection(
		"Queue Worker",
		[]string{conf.RabbitQueues.MainQueue.Queue, conf.RabbitQueues.DelayedQueue.Queue},
		conf.RabbitQueues.MainQueue,
		logger,
	)
	if err != nil {
		logger.Fatal("Can't connect to RabbitMQ", zap.Error(err))
	}

	w := worker.NewWorker(conf, logger, rabbitConsumer, newDB)
	err = w.Run(ctx)
	if err != nil {
		logger.Fatal("Shutdown Service")
		return
	}

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
