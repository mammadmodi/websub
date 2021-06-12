package hub

import (
	"context"
	natsserver "github.com/nats-io/nats-server/test"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func mockNatsHub() (hub *NatsHub, cancel func()) {
	opts := natsserver.DefaultTestOptions
	opts.Port = 8369
	ns := natsserver.RunServer(&opts)

	rc, err := nats.Connect(ns.Addr().String())
	if err != nil || !rc.IsConnected() {
		panic("cannot connect to mock nats server")
	}

	return NewNatsHub(rc, nil, nil), func() { ns.Shutdown() }
}

func TestNewNatsHub(t *testing.T) {
	rc := &nats.Conn{}
	l := logrus.New()
	c := &NatsHubConfig{}
	rh := NewNatsHub(rc, l, c)

	assert.NotNil(t, rh)
	assert.Equal(t, rc, rh.Client)
	assert.Equal(t, l, rh.Logger)
	assert.Equal(t, c, rh.Config)
}

func TestNatsHub(t *testing.T) {
	hub, stop := mockNatsHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		stop()
		cancel()
	}()
	testHubPubSub(ctx, t, hub)
}
