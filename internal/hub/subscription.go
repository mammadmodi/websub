package hub

import (
	"fmt"
	"github.com/go-redis/redis"
	"strings"
)

// Subscription is a struct that holds state of a subscription.
type Subscription struct {
	// Topics is topics string which separated with "," delimeter.
	Topics string
	// Closer is used to closing subscription's underlying connections.
	Closer func()
	// MessageChannel is a go channel that you can receive your messages with that.
	MessageChannel <-chan *redis.Message
}

// Subscribe creates a subscription to topic(or topics) and returns it.
func (r *RedisHub) Subscribe(topics []string) (*Subscription, error) {
	ps := r.Client.Subscribe(topics...)

	s := &Subscription{
		Closer:         func() { _ = ps.Close() },
		MessageChannel: ps.Channel(),
		Topics:         fmt.Sprintf(strings.Join(topics[:], ",")),
	}

	return s, nil
}
