package config

type ExampleTask struct {
	RabbitConnectionName string `mapstructure:"rabbit_connection_name"`
}

type KafkaExampleTask struct {
	KafkaConnectionName string `mapstructure:"kafka_connection_name"`
}
