package hub

import (
	"context"
	"github.com/go-redis/redis"
)

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

func (r RedisHub) Subscribe(ctx context.Context, topic string) (<-chan *redis.Message, func(), error) {
	ps := r.Client.Subscribe(topic)

	ch := ps.Channel()
	closer := func() { _ = ps.Close() }

	return ch, closer, nil
}
