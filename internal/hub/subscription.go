package hub

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

// Subscription is a struct that holds state of a subscription.
type Subscription struct {
	// Topic is topic string.
	Topic string
	// Closer is used to closing subscription's underlying connections.
	Closer func()
	// MessageChannel is a go channel that you can receive your messages with that.
	MessageChannel <-chan *redis.Message
}

// Subscribe creates a subscription to a topic and returns it.
func (r *RedisHub) Subscribe(ctx context.Context, topic string) (*Subscription, error) {
	ps := r.Client.Subscribe(topic)

	s := &Subscription{
		Closer:         func() { _ = ps.Close() },
		MessageChannel: ps.Channel(),
		Topic:          topic,
	}

	return s, nil
}

// BatchSubscribe creates multiple subscriptions with on call.
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
