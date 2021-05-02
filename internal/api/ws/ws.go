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

	// validate request and resolve parameters
	if err := validateRequest(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	// TODO Authenticate and Authorize topics for user
	un := r.URL.Query().Get("username")
	topics := strings.Split(r.URL.Query().Get("topics"), ",")

	// upgrade connection
	wsConn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithField("error", err).WithField("username", un).Error("upgrade error")
		return
	}
	logrus.WithField("username", un).WithField("topics", topics).Info("connection created for user")

	defer func() {
		err := wsConn.Close()
		if err != nil {
			logrus.WithField("error", err).Fatal("error while closing ws connection")
		}
	}()

	// create subscription for user topics
	subs, err := m.Hub.BatchSubscribe(r.Context(), topics)
	if err != nil {
		logrus.WithField("username", un).Info("subscriptions failed for user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	logrus.WithField("username", un).Info("subscriptions created for user")

	defer func() {
		for _, s := range subs {
			s.Closer()
		}
		logrus.WithField("username", un).Info("redis subscriptions closed successfully")
	}()

	go writer(un, wsConn, subs)
	reader(wsConn)
}

func reader(conn *websocket.Conn) {
	// read user sent messages
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			logrus.Println("read:", err)
			break
		}

		logrus.
			WithField("message_type", mt).
			WithField("received_message", string(message)).
			Info("message received from user")
	}
}

func writer(un string, conn *websocket.Conn, subs []*hub.Subscription) {
	for _, s := range subs {
		// pass redis messages to user
		go func(s *hub.Subscription) {
			for msg := range s.MessageChannel {
				logrus.
					WithField("channel", msg.Channel).
					WithField("payload", msg.Payload).
					Info("message received from redis")

				err := conn.WriteMessage(1, []byte(msg.Payload))
				if err != nil {
					logrus.WithField("error", err).Error("error while sending message to user")
					break
				}
			}
		}(s)
		logrus.
			WithField("username", un).
			WithField("topics", s.Topic).
			Info("subscription created")
	}
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
