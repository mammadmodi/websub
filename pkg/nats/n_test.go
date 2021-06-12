package nats

import (
	natsserver "github.com/nats-io/nats-server/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	configs := Configs{
		Address:             "nats://127.0.0.1:8366",
		ConnectTimeout:      10 * time.Second,
		ReconnectWait:       10 * time.Second,
		PingInterval:        10 * time.Second,
		MaxPingsOutstanding: 5,
	}

	// Setup nats test server.
	opts := natsserver.DefaultTestOptions
	opts.Port = 8366
	ns := natsserver.RunServer(&opts)
	// Create connection to test server.
	defer func() {
		ns.Shutdown()
	}()

	nc, err := NewClient(configs)
	if assert.NoError(t, err) {
		assert.True(t, nc.IsConnected())
		assert.Equal(t, configs.ConnectTimeout, nc.Opts.Timeout)
		assert.Equal(t, configs.ReconnectWait, nc.Opts.ReconnectWait)
		assert.Equal(t, configs.PingInterval, nc.Opts.PingInterval)
		assert.Equal(t, configs.MaxPingsOutstanding, nc.Opts.MaxPingsOut)
		assert.Equal(t, configs.MaxPingsOutstanding, nc.Opts.MaxPingsOut)
	}
}
