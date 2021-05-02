package hub

import (
	"context"
	"github.com/go-redis/redis"
)

func (r *RedisHub) Publish(ctx context.Context, channel string, data interface{}) *redis.IntCmd {
	return r.Client.Publish(channel, data)
}
