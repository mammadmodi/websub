package hub

import "github.com/go-redis/redis"

// RedisHub is a redis client wrapper that contains redis pub sub commands.
type RedisHub struct {
	Client redis.UniversalClient
	Config RedisHubConfig
}

// RedisHubConfig is config for RedisHub.
type RedisHubConfig struct{}

// NewRedisHub assigns params to a redis hub object and returns it.
func NewRedisHub(client redis.UniversalClient, config RedisHubConfig) *RedisHub {
	rh := &RedisHub{
		Client: client,
		Config: config,
	}

	return rh
}
