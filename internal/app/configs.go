package app

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/mammadmodi/webis/internal/api/websocket"
	"github.com/mammadmodi/webis/pkg/logger"
	"github.com/mammadmodi/webis/pkg/nats"
	"github.com/mammadmodi/webis/pkg/redis"
	"time"
)

const (
	RedisHub = "redis_hub"
	NatsHub  = "nats_hub"
)

// Configs is struct that contains all configuration of all parts of application
type Configs struct {
	SockHubConfig   websocket.Configuration
	RedisConfigs    redis.Configs
	NatsConfigs     nats.Configs
	LoggingConfigs  logger.Configuration
	HubDriver       string        `default:"redis_hub" split_words:"true"`
	Addr            string        `default:"127.0.0.1"`
	Port            int           `default:"8379"`
	GracefulTimeout time.Duration `default:"15s" split_words:"true"`
}

// NewConfiguration returns a configuration that is loaded with environment variables
func NewConfiguration() (*Configs, error) {
	config := new(Configs)
	err := envconfig.Process("webis", config)
	if err != nil {
		return nil, fmt.Errorf("error while processing global configs from env variables, error: %v", err)
	}

	// loading SockHub configs
	sockHubConfig := websocket.Configuration{}
	err = envconfig.Process("webis_sock", &sockHubConfig)
	if err != nil {
		return nil, fmt.Errorf("error while processing sockhub configs from env variables, error: %v", err)
	}
	config.SockHubConfig = sockHubConfig

	// loading logging configs
	loggingConfig := logger.Configuration{}
	err = envconfig.Process("webis_logging", &loggingConfig)
	if err != nil {
		return nil, fmt.Errorf("error while processing logging configs from env variables, error: %v", err)
	}
	config.LoggingConfigs = loggingConfig

	// loading redis configs
	redisConfigs := redis.Configs{}
	err = envconfig.Process("webis_redis", &redisConfigs)
	if err != nil {
		return nil, fmt.Errorf("error while processing redis client configs from env variables, error: %v", err)
	}
	config.RedisConfigs = redisConfigs

	// loading nats configs
	natsConfigs := nats.Configs{}
	err = envconfig.Process("nats_redis", &natsConfigs)
	if err != nil {
		return nil, fmt.Errorf("error while processing nats client configs from env variables, error: %v", err)
	}
	config.NatsConfigs = natsConfigs

	return config, nil
}
