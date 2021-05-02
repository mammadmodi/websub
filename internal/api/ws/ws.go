package ws

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mammadmodi/webis/internal/source"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

type SocketManager struct {
	Source     *source.RedisSource
	wsUpgrader *websocket.Upgrader
}

func NewSocketManager(redisSource *source.RedisSource) *SocketManager {
	m := &SocketManager{
		Source:     redisSource,
		wsUpgrader: &websocket.Upgrader{},
	}
	return m
}

func (m *SocketManager) Socket(ctx *gin.Context) {
	r := ctx.Request
	w := ctx.Writer
	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logrus.WithField("username", username).Info("request received for user")
	c, err := m.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("upgrade:", err)
		return
	}

	defer func() {
		err := c.Close()
		if err != nil {
			logrus.WithField("error", err).Fatal("error while closing ws connection")
		}
	}()

	logrus.WithField("username", username).Info("connection created for user")

	messageChan, closer, err := m.Source.Subscribe(context.Background(), username)
	if err != nil {
		logrus.WithField("err", err).Error("error while subscription")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer closer()
	logrus.WithField("username", username).Info("subscription created for user")

	// pass redis messages to user
	go func() {
		for msg := range messageChan {
			logrus.
				WithField("channel", msg.Channel).
				WithField("payload", msg.Payload).
				Info("message received from redis")

			err = c.WriteMessage(1, []byte(msg.Payload))
			if err != nil {
				logrus.WithField("error", err).Error("error while sending message to user")
				break
			}
		}
	}()

	// read user sent messages
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		logrus.
			WithField("message_type", mt).
			WithField("received_message", string(message)).
			Info("message received from user")
	}
}
