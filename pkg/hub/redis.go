package hub

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

// RedisHub is a redis client wrapper that contains redis pub sub commands.
type RedisHub struct {
	Client redis.UniversalClient
	Config RedisHubConfig
	Logger *logrus.Logger
}

// RedisHubConfig is config for RedisHub.
type RedisHubConfig struct{}

// NewRedisHub assigns params to a redis hub object and returns it.
func NewRedisHub(client redis.UniversalClient, logger *logrus.Logger, config RedisHubConfig) *RedisHub {
	if logger == nil {
		logger = logrus.New()
		logger.SetOutput(ioutil.Discard)
	}

	rh := &RedisHub{
		Client: client,
		Config: config,
		Logger: logger,
	}

	return rh
}

// Publish publishes a message to a topic.
func (r *RedisHub) Publish(ctx context.Context, message *Message) error {
	cmd := r.Client.Publish(message.Topic, message.Data)
	_, err := cmd.Result()
	return err
}

// Subscribe creates a subscription to topic(or topics) and returns it.
func (r *RedisHub) Subscribe(ctx context.Context, topics ...string) (*Subscription, error) {
	ps := r.Client.Subscribe(topics...)
	msgChannel := make(chan *Message)
	go func() {
		for {
			select {
			case rm := <-ps.Channel():
				msg := &Message{
					Data:  rm.Payload,
					Topic: rm.Pattern,
				}
				msgChannel <- msg
			case <-ctx.Done():
				_ = ps.Close()
				r.Logger.
					WithField("channels", topics).
					Infof("subscription removed from redis")
				return
			}
		}
	}()

	s := &Subscription{
		Topics:         fmt.Sprintf(strings.Join(topics[:], ",")),
		MessageChannel: msgChannel,
	}

	return s, nil
}
