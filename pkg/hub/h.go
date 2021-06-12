package hub

import "context"

// Message is the data type that's been exchanged between hub implementations and .
type Message struct {
	Data  interface{} `json:"data"`
	Topic string      `json:"topic"`
}

// Subscription is a struct that holds state of a subscription.
type Subscription struct {
	// Topics is topics string which separated with "," delimiter.
	Topics string
	// MessageChannel is a go channel that you can receive your messages with that.
	MessageChannel chan *Message
}

// Hub is a messaging channel that implements pub sub exchange pattern.
type Hub interface {
	Publish(ctx context.Context, topic string, data interface{}) (err error)
	Subscribe(ctx context.Context, topics ...string) (*Subscription, error)
}
