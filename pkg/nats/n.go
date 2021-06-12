package nats

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"time"
)

type Configs struct {
	Address             string        `default:"127.0.0.1:4222"`
	ConnectTimeout      time.Duration `split_words:"true" default:"20s"`
	ReconnectWait       time.Duration `split_words:"true" default:"5s"`
	PingInterval        time.Duration `split_words:"true" default:"30s"`
	MaxPingsOutstanding int           `split_words:"true" default:"5"`
	ClusterID           string        `split_words:"true" default:"webis_cluster"`
	ClientID            string        `split_words:"true" default:"webis_client"`
}

func NewClient(configs Configs) (natsClient *nats.Conn, err error) {
	conn, err := nats.Connect(configs.Address,
		nats.Timeout(configs.ConnectTimeout),
		nats.PingInterval(configs.PingInterval),
		nats.RetryOnFailedConnect(true),
		nats.ReconnectWait(configs.ReconnectWait),
		nats.MaxPingsOutstanding(configs.MaxPingsOutstanding),
	)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to nats server, error: %s", err.Error())
	}

	return conn, nil
}
