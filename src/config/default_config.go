package config

import (
	"time"
)

var RabbitMQDefaults = map[string]interface{}{
	"main_queue": RabbitMQ{
		AmqpURI: "",
	},
	"delayed_queue": RabbitMQ{
		AmqpURI: "",
	},
}

// //var DefaultDRS = map[string]interface{}{
// //	"url":      "localhost",
// //	"usewhois": true,
// //}

var DefaultConfig = map[string]map[string]interface{}{
	// "rabbit_queues": RabbitMQDefaults,
	// "drs":      DefaultDRS,
}

func (c *Config) loadEnv() {
	c.Logging.RemoteAddr = c.envVariables("GRAYLOG_ADDRESS")
	c.Logging.RemoteProtocol = c.envVariables("GRAYLOG_PROTOCOL")

	c.MongoDB.DatabaseURL = c.envVariables("MONGODB_URL")

	c.RabbitQueues.DelayedQueue.AmqpURI = c.envVariables("RABBIT_QUEUES_DELAYED_QUEUE_URI")
	c.RabbitQueues.MainQueue.AmqpURI = c.envVariables("RABBIT_QUEUES_MAIN_QUEUE_URI")
}

func InitConfig() Config {
	c := Config{}

	logger := &Logging{LogLevel: "debug"}

	c.RabbitQueues = &RabbitQueues{}

	mainQueue := &RabbitMQ{
		ConnName:     "Tasks_Test_Connection",
		AmqpURI:      "amqp://soltelab:labpass@localhost:5672/",
		Queue:        "defoult-queue",
		Consumer:     "defoult-consumer",
		Exchange:     "main-exchange",
		ExchangeType: "direct",
		BindingKey:   "main-queue",
		ConsumerTag:  "lab-test-consumer",
		Verbose:      false,
		AutoAck:      false,
		Durable:      true,
		Qos:          1,
		ExchangeArgs: ExchangeArgs{},
	}

	delayedQueue := &RabbitMQ{
		ConnName:     "Tasks_Test_Delayed_Connection",
		AmqpURI:      "amqp://soltelab:labpass@localhost:5672/",
		Queue:        "delayed-queue",
		Consumer:     "defoult-consumer",
		Exchange:     "delayed-exchange",
		ExchangeType: "x-delayed-message",
		BindingKey:   "delayed-queue",
		ConsumerTag:  "lab-test-consumer",
		Verbose:      false,
		AutoAck:      false,
		Durable:      true,
		Qos:          1,
		ExchangeArgs: ExchangeArgs{
			Key:   "x-delayed-type",
			Value: "direct",
			Delay: 5 * time.Second,
		},
	}

	mongo := &MongoDB{
		DatabaseURL:       "mongodb://dbtest:supra**@localhost:27017/?authSource=admin",
		MaxPoolSize:       20,
		ConnectTimeout:    30 * time.Second,
		DatabaseName:      "developmet_db",
		IndicesCollection: "shared_collection",
	}

	indices := make(map[string]string)
	indices["main_index"] = "main-development"
	indices["log_index"] = "log-developmen"

	main := &Elastic{
		URLS:    []string{"http://localhost:9200"},
		Indices: indices,
		Auth:    BasicAuth{},
	}

	c.ElasticSearch = &ElasticSearch{
		Main: main,
	}

	sch := &Scheduler{
		Tasks:  [][3]string{{"1h", "test_task", ""}},
		Jitter: 0,
	}

	c.MongoDB = mongo
	c.Logging = logger

	c.RabbitQueues = &RabbitQueues{
		MainQueue:    mainQueue,
		DelayedQueue: delayedQueue,
	}

	c.Tasks = &Tasks{}
	c.Scheduler = sch

	return c
}
