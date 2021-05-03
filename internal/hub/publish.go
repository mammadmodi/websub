package hub

import (
	"context"
	"github.com/go-redis/redis"
)

// Publish publishes a message to a topic.
func (r *RedisHub) Publish(ctx context.Context, channel string, data interface{}) *redis.IntCmd {
	return r.Client.Publish(channel, data)
}
