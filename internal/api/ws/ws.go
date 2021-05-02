package ws

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mammadmodi/webis/internal/hub"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

type SockHub struct {
	Hub      *hub.RedisHub
	upgrader *websocket.Upgrader
}

func NewSockHub(redisHub *hub.RedisHub) *SockHub {
	m := &SockHub{
		Hub:      redisHub,
		upgrader: &websocket.Upgrader{},
	}
	return m
}

func (m *SockHub) Socket(ctx *gin.Context) {
	r := ctx.Request
	w := ctx.Writer
	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logrus.WithField("username", username).Info("request received for user")
	c, err := m.upgrader.Upgrade(w, r, nil)
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

	messageChan, closer, err := m.Hub.Subscribe(context.Background(), fmt.Sprintf("events/%s", username))
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
