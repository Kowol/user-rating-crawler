package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AMQP     AMQP     `required:"true"`
	Database Database `required:"true"`
	GRPC     GRPC     `required:"true"`
	Crawler  Crawler  `required:"true"`
}

type AMQP struct {
	URL          string `required:"true" envconfig:"AMQP_URL"`
	QueueName    string `required:"true" envconfig:"AMQP_QUEUE_NAME" default:"channel_crawler"`
	ExchangeName string `required:"true" envconfig:"AMQP_EXCHANGE_NAME" default:"urls"`
	RoutingKey   string `required:"true" envconfig:"AMQP_ROUTING_KEY" default:"channel_url"`
}

type Database struct {
	DSN          string `required:"true" envconfig:"DATABASE_DSN"`
	DatabaseName string `required:"true" envconfig:"DATABASE_DB_NAME" default:"crawler"`
}

type GRPC struct {
	ServerPort int `required:"true" envconfig:"GRPC_SERVER_PORT"`
}

type Crawler struct {
	WorkersAmount int `required:"true" envconfig:"CRAWLER_WORKERS_AMOUNT" default:"5"`
}

func ParseConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
