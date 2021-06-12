package hub

import (
	"context"
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

func TestRedisHubPubSub(t *testing.T) {
	redisHub, stop := mockRedisHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		stop()
		cancel()
	}()
	testHubPubSub(ctx, t, redisHub)
}
