package hub

import (
	"context"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func mockRedisHub() (hub *RedisHub, cancel func()) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	rc := redis.NewClient(&redis.Options{Addr: s.Addr()})

	return NewRedisHub(rc, nil, nil), func() { s.Close() }
}

func TestNewRedisHub(t *testing.T) {
	rc := &redis.Client{}
	l := logrus.New()
	c := &RedisHubConfig{}
	rh := NewRedisHub(rc, l, c)

	assert.NotNil(t, rh)
	assert.Equal(t, rc, rh.Client)
	assert.Equal(t, l, rh.Logger)
	assert.Equal(t, c, rh.Config)
}

func TestRedisHub(t *testing.T) {
	redisHub, stop := mockRedisHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		stop()
		cancel()
	}()

	// Creating subscription to two different topics.
	sub, err := redisHub.Subscribe(ctx, "topic1", "topic2")
	assert.NoError(t, err)
	assert.NotNil(t, sub.MessageChannel)
	assert.Equal(t, sub.Topics, "topic1,topic2")

	// Launching a go routine to receive messages from subscribed channels.
	var wg sync.WaitGroup
	wg.Add(2)
	var receivedMessages []Message
	go func() {
		for {
			select {
			case msg := <-sub.MessageChannel:
				receivedMessages = append(receivedMessages, *msg)
				wg.Done()
			case <-ctx.Done():
				return
			}
		}
	}()

	publishingMessages := []Message{
		{
			Data:  `{"key": "value"}`,
			Topic: "topic1",
		},
		{
			Data:  `{"secondKey": "secondValue"}`,
			Topic: "topic2",
		},
	}
	// Publishing messages to redis.
	for _, m := range publishingMessages {
		err := redisHub.Publish(ctx, m.Topic, m.Data)
		assert.NoError(t, err)
	}

	// Waits til receive published messages in subscription go routine.
	wg.Wait()

	assert.EqualValues(t, publishingMessages, receivedMessages)
}
