package hub

import "github.com/go-redis/redis"

type RedisHub struct {
	Client redis.UniversalClient
	Config RedisHubConfig
}

type RedisHubConfig struct{}

func NewRedisHub(client redis.UniversalClient, config RedisHubConfig) *RedisHub {
	rh := &RedisHub{
		Client: client,
		Config: config,
	}

	return rh
}
