// Package websocket make a websocket server which tunnels connections to
// other services and users with a backend pubsub messaging tool like redis.
package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/mammadmodi/webis/internal/hub"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// Configuration is used in SockHub method set.
type Configuration struct {
	// PingInterval is interval of ping messages that will be sent to client.
	PingInterval time.Duration `default:"25s" split_words:"true"`
	// PongWait is duration that server should wait for a ping response.
	PongWait time.Duration `default:"30s" split_words:"true"`
	// WriteWait is a timeout for writing messages to client.
	WriteWait time.Duration `default:"20s" split_words:"true"`
	// ReadLimit is maximum size of messages(in Bytes) that is received from user.
	ReadLimit int64 `default:"4096" split_words:"true"`
}

// SockHub tunnels websocket messages(in and out) to a pubsub hub.
type SockHub struct {
	// Hub is a core pubsub driver(e.g. RedisHub) that is used to tunneling messages.
	Hub    hub.Hub
	Config Configuration

	logger   *logrus.Logger
	upgrader *websocket.Upgrader
}

// NewSockHub creates a SockHub object.
func NewSockHub(config Configuration, hub hub.Hub, logger *logrus.Logger) *SockHub {
	m := &SockHub{
		Hub:    hub,
		Config: config,
		logger: logger,
		upgrader: &websocket.Upgrader{
			// TODO you should not ignore origin check in production.
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
	return m
}
