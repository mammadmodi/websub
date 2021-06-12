package redis_test

import (
	"github.com/go-redis/redis"
	. "github.com/mammadmodi/websub/pkg/redis"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	configs := Configs{
		Address:            "127.0.0.1:6379",
		MasterAddress:      "127.0.0.2:6379",
		SlaveAddress:       "127.0.0.3:6379",
		DB:                 5,
		Password:           "12345678",
		PoolSize:           10,
		MaxRetries:         2,
		DialTimeout:        15 * time.Second,
		ReadTimeout:        4 * time.Second,
		WriteTimeout:       4 * time.Second,
		PoolTimeout:        60 * time.Second,
		MinRetryBackoff:    15 * time.Millisecond,
		MaxRetryBackoff:    512 * time.Millisecond,
		IdleTimeout:        60 * time.Second,
		IdleCheckFrequency: 30,
		ReadOnly:           true,
		RouteRandomly:      true,
	}

	t.Run("testing new client for single node redis", func(t *testing.T) {
		configs.Mode = SingleNode
		c, err := NewClient(configs)
		assert.NoError(t, err)
		assert.NotNil(t, c)
		sc, ok := c.(*redis.Client)
		if !assert.True(t, ok) {
			t.Error("redis client's type must be *redis.Client")
			return
		}
		assert.Equal(t, configs.Address, sc.Options().Addr)
		assert.Equal(t, configs.DB, sc.Options().DB)
		assert.Equal(t, configs.Password, sc.Options().Password)
		assert.Equal(t, configs.PoolSize, sc.Options().PoolSize)
		assert.Equal(t, configs.MaxRetries, sc.Options().MaxRetries)
		assert.Equal(t, configs.DialTimeout, sc.Options().DialTimeout)
		assert.Equal(t, configs.ReadTimeout, sc.Options().ReadTimeout)
		assert.Equal(t, configs.WriteTimeout, sc.Options().WriteTimeout)
		assert.Equal(t, configs.PoolTimeout, sc.Options().PoolTimeout)
		assert.Equal(t, configs.MinRetryBackoff, sc.Options().MinRetryBackoff)
		assert.Equal(t, configs.MaxRetryBackoff, sc.Options().MaxRetryBackoff)
		assert.Equal(t, configs.IdleTimeout, sc.Options().IdleTimeout)
		assert.Equal(t, configs.IdleCheckFrequency, sc.Options().IdleCheckFrequency)
	})

	t.Run("testing new client for cluster redis", func(t *testing.T) {
		configs.Mode = Cluster
		c, err := NewClient(configs)
		assert.NoError(t, err)
		assert.NotNil(t, c)
		cc, ok := c.(*redis.ClusterClient)
		if !assert.True(t, ok) {
			t.Error("redis client's type must be *redis.ClusterClient")
			return
		}
		assert.Equal(t, []string{configs.MasterAddress, configs.SlaveAddress}, cc.Options().Addrs)
		assert.Equal(t, configs.Password, cc.Options().Password)
		assert.Equal(t, configs.PoolSize, cc.Options().PoolSize)
		assert.Equal(t, configs.MaxRetries, cc.Options().MaxRetries)
		assert.Equal(t, configs.DialTimeout, cc.Options().DialTimeout)
		assert.Equal(t, configs.ReadTimeout, cc.Options().ReadTimeout)
		assert.Equal(t, configs.WriteTimeout, cc.Options().WriteTimeout)
		assert.Equal(t, configs.PoolTimeout, cc.Options().PoolTimeout)
		assert.Equal(t, configs.MinRetryBackoff, cc.Options().MinRetryBackoff)
		assert.Equal(t, configs.MaxRetryBackoff, cc.Options().MaxRetryBackoff)
		assert.Equal(t, configs.IdleTimeout, cc.Options().IdleTimeout)
		assert.Equal(t, configs.IdleCheckFrequency, cc.Options().IdleCheckFrequency)
		assert.Equal(t, configs.ReadOnly, cc.Options().ReadOnly)
		assert.Equal(t, configs.RouteRandomly, cc.Options().RouteRandomly)
	})

	t.Run("testing new client for invalid redis mode", func(t *testing.T) {
		configs.Mode = "invalid_mode"
		c, err := NewClient(configs)
		assert.Error(t, err)
		assert.Nil(t, c)
	})
}
