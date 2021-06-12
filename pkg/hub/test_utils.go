package hub

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func testHubPubSub(ctx context.Context, t *testing.T, hub Hub) {
	// Creating subscription to two different topics.
	sub, err := hub.Subscribe(ctx, "topic1", "topic2")
	assert.NoError(t, err)
	assert.NotNil(t, sub.MessageChannel)
	assert.Equal(t, sub.Topics, "topic1,topic2")

	// Launching a go routine to receive messages from subscribed channels.
	var wg sync.WaitGroup
	wg.Add(2)
	receivedMessages := map[string]Message{}
	go func() {
		for {
			select {
			case msg := <-sub.MessageChannel:
				receivedMessages[msg.Topic] = *msg
				wg.Done()
			case <-ctx.Done():
				return
			}
		}
	}()

	publishingMessages := map[string]Message{
		"topic1": {
			Data:  `{"key": "value"}`,
			Topic: "topic1",
		},
		"topic2": {
			Data:  `{"secondKey": "secondValue"}`,
			Topic: "topic2",
		},
	}
	// Publishing messages to redis.
	for _, m := range publishingMessages {
		err := hub.Publish(ctx, m.Topic, m.Data)
		assert.NoError(t, err)
	}

	// Waits til receive published messages in subscription go routine.
	wg.Wait()

	assert.EqualValues(t, publishingMessages, receivedMessages)
}
