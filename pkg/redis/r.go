// Package redis is an abstraction layer in top of "github.com/go-redis/redis" package.
package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

// Mode is the type our redis server and can be cluster, single node ... .
type Mode string

const (
	Cluster    Mode = "cluster"
	SingleNode Mode = "single_node"
)

type Configs struct {
	Mode               Mode          `split_words:"true" default:"single_node"`
	Address            string        `split_words:"true" default:"127.0.0.1:6379"`
	MasterAddress      string        `split_words:"true" default:"127.0.0.1:6379"`
	SlaveAddress       string        `split_words:"true" default:"127.0.0.1:6379"`
	DB                 int           `split_words:"true" default:"0"`
	Password           string        `split_words:"true" default:""`
	PoolSize           int           `split_words:"true" default:"10"`
	MaxRetries         int           `split_words:"true" default:"1"`
	DialTimeout        time.Duration `split_words:"true" default:"5s"`
	ReadTimeout        time.Duration `split_words:"true" default:"250ms"`
	WriteTimeout       time.Duration `split_words:"true" default:"400ms"`
	PoolTimeout        time.Duration `split_words:"true" default:"4s"`
	MinRetryBackoff    time.Duration `split_words:"true" default:"20ms"`
	MaxRetryBackoff    time.Duration `split_words:"true" default:"80ms"`
	IdleTimeout        time.Duration `split_words:"true" default:"60s"`
	IdleCheckFrequency time.Duration `split_words:"true" default:"60s"`
	ReadOnly           bool          `split_words:"true" default:"true"`
	RouteRandomly      bool          `split_words:"true" default:"false"`
}

// NewClient is a factory function that creates and initializes a proper redis client.
func NewClient(configs Configs) (redisClient redis.UniversalClient, err error) {
	switch configs.Mode {
	case SingleNode:
		opts := &redis.Options{
			Addr:               configs.Address,
			Password:           configs.Password,
			DB:                 configs.DB,
			MaxRetries:         configs.MaxRetries,
			MinRetryBackoff:    configs.MinRetryBackoff,
			MaxRetryBackoff:    configs.MaxRetryBackoff,
			DialTimeout:        configs.DialTimeout,
			ReadTimeout:        configs.ReadTimeout,
			WriteTimeout:       configs.WriteTimeout,
			PoolSize:           configs.PoolSize,
			PoolTimeout:        configs.PoolTimeout,
			IdleTimeout:        configs.IdleTimeout,
			IdleCheckFrequency: configs.IdleCheckFrequency,
		}
		redisClient = redis.NewClient(opts)
	case Cluster:
		opts := &redis.ClusterOptions{
			Addrs:              []string{configs.MasterAddress, configs.SlaveAddress},
			Password:           configs.Password,
			MaxRetries:         configs.MaxRetries,
			MinRetryBackoff:    configs.MinRetryBackoff,
			MaxRetryBackoff:    configs.MaxRetryBackoff,
			DialTimeout:        configs.DialTimeout,
			ReadTimeout:        configs.ReadTimeout,
			WriteTimeout:       configs.WriteTimeout,
			PoolSize:           configs.PoolSize,
			PoolTimeout:        configs.PoolTimeout,
			IdleTimeout:        configs.IdleTimeout,
			IdleCheckFrequency: configs.IdleCheckFrequency,
			ReadOnly:           configs.ReadOnly,
			RouteRandomly:      configs.RouteRandomly,
		}
		redisClient = redis.NewClusterClient(opts)
	default:
		return nil, fmt.Errorf("'%s' is not a valid mode", configs.Mode)
	}
	return
}
