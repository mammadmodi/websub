package hub

import (
	"context"
	"github.com/go-redis/redis"
)

type Subscription struct {
	Topic          string
	Closer         func()
	MessageChannel <-chan *redis.Message
}

func (r RedisHub) Subscribe(ctx context.Context, topic string) (*Subscription, error) {
	ps := r.Client.Subscribe(topic)

	s := &Subscription{
		Closer:         func() { _ = ps.Close() },
		MessageChannel: ps.Channel(),
		Topic:          topic,
	}

	return s, nil
}
