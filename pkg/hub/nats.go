package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

// NatsHub is a nats client wrapper that contains nats pub sub commands.
type NatsHub struct {
	Client *nats.Conn
	Config *NatsHubConfig
	Logger *logrus.Logger
}

// NatsHubConfig is config for NatsHub.
type NatsHubConfig struct{}

// NewNatsHub assigns params to a nats hub object and returns it.
func NewNatsHub(client *nats.Conn, logger *logrus.Logger, config *NatsHubConfig) *NatsHub {
	if logger == nil {
		logger = logrus.New()
		logger.SetOutput(ioutil.Discard)
	}

	rh := &NatsHub{
		Client: client,
		Config: config,
		Logger: logger,
	}

	return rh
}

// Publish publishes a message to a topic.
func (n *NatsHub) Publish(_ context.Context, topic string, data interface{}) (err error) {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error while marshalling message data, error : %s", err.Error())
	}
	n.Logger.WithField("subject", topic).Debug("successfully published to nats")
	return n.Client.Publish(topic, b)
}

// Subscribe creates a subscription to topic(or topics) and returns it.
func (n *NatsHub) Subscribe(ctx context.Context, topics ...string) (*Subscription, error) {
	msgChannel := make(chan *Message)
	subs := make([]*nats.Subscription, len(topics), len(topics))
	for _, t := range topics {
		subject := t
		s, err := n.Client.Subscribe(t, func(msg *nats.Msg) {
			n.Logger.WithField("subject", subject).Debug("message received by nats")
			hm := &Message{
				Data:  string(msg.Data),
				Topic: subject,
			}
			msgChannel <- hm
		})
		if err != nil {
			return nil, fmt.Errorf("error while creating nats subscription to %s, error: %s", subject, err.Error())
		}
		subs = append(subs, s)
	}

	go func() {
		<-ctx.Done()
		n.Logger.WithField("subject", topics).Debug("context is done for nats subscriptions")

		for _, s := range subs {
			_ = s.Unsubscribe()
		}
	}()

	s := &Subscription{
		Topics:         fmt.Sprintf(strings.Join(topics[:], ",")),
		MessageChannel: msgChannel,
	}

	return s, nil
}
