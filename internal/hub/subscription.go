package hub

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type Subscription struct {
	Topic          string
	Closer         func()
	MessageChannel <-chan *redis.Message
}

func (r *RedisHub) Subscribe(ctx context.Context, topic string) (*Subscription, error) {
	ps := r.Client.Subscribe(topic)

	s := &Subscription{
		Closer:         func() { _ = ps.Close() },
		MessageChannel: ps.Channel(),
		Topic:          topic,
	}

	return s, nil
}

func (r *RedisHub) BatchSubscribe(ctx context.Context, topics []string) ([]*Subscription, error) {
	subs := make([]*Subscription, len(topics))

	for i, t := range topics {
		s, err := r.Subscribe(ctx, t)
		// TODO UnSub prev subscription when one error occurs
		if err != nil {
			logrus.WithField("err", err).WithField("topic", t).Error("error while subscription")
			return nil, fmt.Errorf("error while subscription to topic %s", t)
		}
		subs[i] = s
	}
	return subs, nil
}
