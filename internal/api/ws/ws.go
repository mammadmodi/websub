package ws

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mammadmodi/webis/internal/hub"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
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

	if err := validateRequest(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// TODO Authenticate and Authorize topics for user
	un := r.URL.Query().Get("username")
	topics := strings.Split(r.URL.Query().Get("topics"), ",")

	logrus.WithField("username", un).WithField("topics", topics).Info("request received for user")
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

	logrus.WithField("username", un).Info("connection created for user")

	subs := make([]*hub.Subscription, len(topics))

	for i, t := range topics {
		s, err := m.Hub.Subscribe(r.Context(), t)
		// TODO UnSub prev subscription when one error occurs
		if err != nil {
			logrus.WithField("err", err).Error("error while subscription")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		subs[i] = s
	}

	defer func() {
		for _, s := range subs {
			s.Closer()
		}
	}()

	logrus.WithField("username", un).Info("subscriptions created for user")

	for _, s := range subs {
		// pass redis messages to user
		go func(s *hub.Subscription) {
			for msg := range s.MessageChannel {
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
		}(s)
		logrus.WithField("username", un).WithField("username", s.Topic).Info("subscription created")
	}

	go func() {
		// read user sent messages
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				logrus.Println("read:", err)
				break
			}

			logrus.
				WithField("message_type", mt).
				WithField("received_message", string(message)).
				Info("message received from user")
		}
	}()

	// TODO Handle close in a proper way
	<-r.Context().Done()
	logrus.WithField("username", un).Info("connection closed for user")
}

func validateRequest(req *http.Request) error {
	username := req.URL.Query().Get("username")
	if username == "" {
		return errors.New("username cannot be empty")
	}

	topics := req.URL.Query().Get("topics")
	if topics == "" {
		return errors.New("topics cannot be empty")
	}

	return nil
}
