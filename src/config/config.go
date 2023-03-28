package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Environment   string
	ElasticSearch *ElasticSearch `mapstructure:"elastic"`
	Server        *HTTPServer    `mapstructure:"http_server"`
	OraAPI        *OraAPI        `mapstructure:"ora_api"`
	Kafka         *Kafka         `mapstructure:"v1"`
	Logging       *Logging       `mapstructure:"logging"`
	MongoDB       *MongoDB       `mapstructure:"mongo"`
	RabbitQueues  *RabbitQueues  `mapstructure:"rabbit_queues"`
	Scheduler     *Scheduler     `mapstructure:"scheduler"`
	Tasks         *Tasks         `mapstructure:"tasks"`
	Worker        Worker         `mapstructure:"worker"`
}

type Tasks struct {
	ExampleTask      `mapstructure:"example_task"`
	KafkaExampleTask `mapstructure:"kafka_example_task"`
}

type HTTPServer struct {
	Port int `mapstructure:"apiPort"`
}

type Scheduler struct {
	Tasks  [][3]string   `mapstructure:"tasks"`
	Jitter time.Duration `mapstructure:"jitter"`
}

type RabbitQueues struct {
	MainQueue    *RabbitMQ `mapstructure:"main_queue"`
	DelayedQueue *RabbitMQ `mapstructure:"delayed_queue"`
}

type Kafka struct {
	ConnectionName string        `mapstructure:"connection_name"`
	Brokers        string        `mapstructure:"brokers"`
	Topic          string        `mapstructure:"topic"`
	TimeOut        time.Duration `mapstructure:"timeout"`
	Partitions     int           `mapstructure:"partitions"`
	PushTimeout    time.Duration `mapstructure:"push_timeout"`
}

type RabbitMQ struct {
	ConnName     string       `mapstructure:"conn_name"`
	AmqpURI      string       `mapstructure:"uri"`
	Queue        string       `mapstructure:"queue"`
	Consumer     string       `mapstructure:"consumer"`
	Exchange     string       `mapstructure:"exchange"`
	ExchangeType string       `mapstructure:"exchange_type"`
	ExchangeArgs ExchangeArgs `mapstructure:"exchange_args"`
	BindingKey   string       `mapstructure:"binding_key"`
	ConsumerTag  string       `mapstructure:"consumer_tag"`
	Verbose      bool         `mapstructure:"verbose"`
	AutoAck      bool         `mapstructure:"auto_ack"`
	Durable      bool         `mapstructure:"durable"`
	Qos          int          `mapstructure:"qos"`
}

type ExchangeArgs struct {
	Key   string        `mapstructure:"arg_key"`
	Value string        `mapstructure:"arg_value"`
	Delay time.Duration `mapstructure:"delay"`
}

type OraAPI struct {
	URL      string `mapstructure:"url"`
	Instance string `mapstructure:"instance"`
}

type MongoDB struct {
	DatabaseURL    string        `mapstructure:"database_url"`
	MaxPoolSize    int           `mapstructure:"max_pool_size"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
	DatabaseName   string        `mapstructure:"database_name"`
	Collection     string        `mapstructure:"collection"`
}

type ElasticSearch struct {
	Main *Elastic `mapstructure:"main"`
}

type Elastic struct {
	URLS    []string          `mapstructure:"urls"`
	Indices map[string]string `mapstructure:"indices"`
	Auth    BasicAuth
	Client  *elasticsearch.Client
}

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Logging struct {
	LogLevel       string `mapstructure:"loglevel"`
	RemoteAddr     string `mapstructure:"remote_address"`
	RemoteProtocol string `mapstructure:"remote_protocol"`
	RemotePort     int    `mapstructure:"remote_port"`
}

type Worker struct {
	BunchSize      int           `mapstructure:"bunch_size"`
	TickerInterval time.Duration `mapstructure:"ticker_interval"`
	TaskTimeout    time.Duration `mapstructure:"task_timeout"`
	TaskConnDelay  time.Duration `mapstructure:"task_conn_delay"`
}

func loadDefaults(v *viper.Viper) {
	for key, val := range DefaultConfig {
		v.SetDefault(key, val)
	}
}

func LoadConf(env string) (Config, error) {
	var c Config

	var confFileName string
	if env == "prod" {
		confFileName = "production"
		err := c.readEnvironment(".env")
		if err != nil {
			return Config{}, err
		}
	} else {
		confFileName = "development"
		err := c.readEnvironment("../.env")
		if err != nil {
			return Config{}, err
		}
	}

	v := viper.New()
	v.SetConfigName(confFileName)
	v.AddConfigPath("../")
	if err := v.ReadInConfig(); err != nil {
		return Config{}, err
	}

	v.AddConfigPath("../charts")
	files := c.prepareConfingFiles()

	for _, file := range files {
		v.SetConfigName(file)
		if err := v.MergeInConfig(); err != nil {
			return Config{}, err
		}
	}

	loadDefaults(v)

	if err := v.Unmarshal(&c); err != nil {
		return Config{}, err
	}

	// Setup environment. Used for adjusting logic for different instances(dev & prod)
	c.Environment = env
	c.loadEnv()
	return c, nil
}

// envVariables assigns variables from Environment
func (c *Config) envVariables(key string) string {
	return os.Getenv(key)
}

// readEnvironment reads the first existing env file from the list
func (c *Config) readEnvironment(files ...string) error {
	for _, f := range files {
		err := godotenv.Load(f)
		if err == nil {
			return nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return nil
}

func (c *Config) prepareConfingFiles() []string {
	var tasksConfigFiles []string

	matches, err := filepath.Glob("../charts/*.toml")
	if err != nil {
		return nil
	}

	for _, m := range matches {
		filePath := strings.Split(m, ".")
		file := strings.Split(filePath[2], "\\")
		tasksConfigFiles = append(tasksConfigFiles, file[2])
	}

	return tasksConfigFiles
}
