package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
	"github.com/mammadmodi/webis/internal/hub"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

type Configuration struct {
	PingInterval time.Duration `default:"4s" split_words:"true"`
	PongWait     time.Duration `default:"6s" split_words:"true"`
	WriteWait    time.Duration `default:"4s" split_words:"true"`
	ReadLimit    int64         `default:"4096" split_words:"true"`
}

// GetConfigFromEnv tries to generate configuration from related
// environment variables with power of "kelseyhightower" library.
func GetConfigFromEnv(prefix string) (Configuration, error) {
	config := Configuration{}
	if err := envconfig.Process(prefix, &config); err != nil {
		return config, fmt.Errorf("error while loading configs from env variables, error: %v", err)
	}

	return config, nil
}

type SockHub struct {
	Hub    *hub.RedisHub
	Config Configuration

	upgrader *websocket.Upgrader
}

type ClientMessage struct {
	Body  string `json:"body"`
	Topic string `json:"topic"`
}

func NewSockHub(config Configuration, redisHub *hub.RedisHub) *SockHub {
	m := &SockHub{
		Hub:      redisHub,
		Config:   config,
		upgrader: &websocket.Upgrader{},
	}
	return m
}

func (h *SockHub) Socket(ctx *gin.Context) {
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
	wsConn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithField("error", err).WithField("username", un).Error("upgrade error")
		return
	}
	logrus.WithField("username", un).WithField("topics", topics).Info("connection created for user")

	defer func() {
		err := wsConn.Close()
		if err != nil {
			logrus.WithField("error", err).Error("error while closing ws connection")
		}
		logrus.WithField("username", un).Info("socket connection closed")
	}()

	// create subscription for user topics
	subs, err := h.Hub.BatchSubscribe(r.Context(), topics)
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

	pingTicker := time.NewTicker(h.Config.PingInterval)
	defer pingTicker.Stop()

	go h.writer(pingTicker, un, wsConn, subs)
	h.reader(r.Context(), un, wsConn)
}

func (h *SockHub) reader(ctx context.Context, username string, conn *websocket.Conn) {
	conn.SetReadLimit(h.Config.ReadLimit)

	if err := conn.SetReadDeadline(time.Now().Add(h.Config.PongWait)); err != nil {
		logrus.WithField("error", err.Error()).Error("error while setting read deadline")
		return
	}

	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(h.Config.PongWait)); err != nil {
			return fmt.Errorf("error while setting read deadline, error: %s", err.Error())
		}
		return nil
	})

	// read user sent messages
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		cm := &ClientMessage{}
		if err = json.Unmarshal(message, cm); err != nil {
			logrus.
				WithField("username", username).
				WithField("type", mt).
				WithField("payload", cm).
				Info("invalid error from usre")
			continue
		}
		logrus.
			WithField("username", username).
			WithField("type", mt).
			WithField("payload", cm).
			Info("message received from user")
		h.Hub.Publish(ctx, cm.Topic, cm.Body)
	}
}

func (h *SockHub) writer(pingTicker *time.Ticker, un string, conn *websocket.Conn, subs []*hub.Subscription) {
	go func() {
		defer func() {
			pingTicker.Stop()
		}()

		for {
			<-pingTicker.C
			logrus.WithField("username", un).Debug("writing ping message")
			if err := conn.SetWriteDeadline(time.Now().Add(h.Config.WriteWait)); err != nil {
				logrus.WithField("error", err.Error()).Error("error while setting write deadline")
				return
			}

			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				logrus.WithField("error", err.Error()).Error("error while sending ping message")
				return
			}
			logrus.WithField("username", un).Debug("ping sent")
		}
	}()

	for _, s := range subs {
		// pass redis messages to user
		go func(s *hub.Subscription) {
			logrus.WithField("topic", s.Topic).Debug("listening to message channel")
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
			logrus.WithField("topic", s.Topic).Debug("message channel closed")
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
