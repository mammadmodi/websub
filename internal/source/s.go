package source

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
)

type RedisSource struct {
	Client redis.UniversalClient
}

func (r RedisSource) Subscribe(ctx context.Context, username string) (<-chan *redis.Message, func(), error) {
	ps := r.Client.Subscribe(fmt.Sprintf("events/%s", username))

	ch := ps.Channel()
	closer := func() { _ = ps.Close() }

	return ch, closer, nil
}
