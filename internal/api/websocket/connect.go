package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/mammadmodi/webis/internal/hub"
	"net/http"
	"strings"
	"time"
)

// ClientMessage is structure of messages that will be received from user.
type ClientMessage struct {
	Body  string `json:"body"`
	Topic string `json:"topic"`
}

// Connect is a http handler that in first upgrades protocol to Websocket Protocol and
// then creates subscriptions to topics which user is requested.
func (h *SockHub) Connect(w http.ResponseWriter, r *http.Request) {
	// Validate request and resolve parameters
	if err := validateRequest(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	// TODO Authenticate and Authorize topic accesses for user.
	un := r.URL.Query().Get("username")
	topics := strings.Split(r.URL.Query().Get("topics"), ",")

	// Upgrade http connection to websocket and configure connection.
	wsConn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithField("error", err).WithField("username", un).Error("upgrade error")
		return
	}
	wsConn.SetReadLimit(h.Config.ReadLimit)
	if err := wsConn.SetReadDeadline(time.Now().Add(h.Config.PongWait)); err != nil {
		h.logger.WithField("error", err.Error()).Error("error while setting read deadline")
		return
	}
	wsConn.SetPongHandler(func(string) error {
		if err := wsConn.SetReadDeadline(time.Now().Add(h.Config.PongWait)); err != nil {
			return fmt.Errorf("error while setting read deadline, error: %s", err.Error())
		}
		h.logger.WithField("username", un).Debug("pong received")
		return nil
	})
	h.logger.WithField("username", un).WithField("topics", topics).Info("connection created for user")

	// Schedule ws connection close at the end.
	defer func() {
		err := wsConn.Close()
		if err != nil {
			h.logger.WithField("error", err).Error("error while closing ws connection")
		}
		h.logger.WithField("username", un).Info("socket connection closed")
	}()

	// Create hub subscriptions for user topics.
	subs, err := h.Hub.BatchSubscribe(r.Context(), topics)
	if err != nil {
		h.logger.WithField("username", un).Info("subscriptions failed for user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.logger.WithField("username", un).Info("hub subscriptions created for user")

	// Schedule hub unsubscribe at the end.
	defer func() {
		for _, s := range subs {
			s.Closer()
		}
		h.logger.WithField("username", un).Info("hub subscriptions closed successfully")
	}()

	// Launch a ws pinger in background.
	pingTicker := time.NewTicker(h.Config.PingInterval)
	defer pingTicker.Stop()
	go h.pingOnTick(un, pingTicker, wsConn)

	h.writer(un, wsConn, subs)
	h.reader(r.Context(), un, wsConn)
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

// pingOnTick sends a ping message to user when receives a signal from ping ticker.
func (h *SockHub) pingOnTick(un string, pingTicker *time.Ticker, conn *websocket.Conn) {
	for {
		<-pingTicker.C
		h.logger.WithField("username", un).Debug("writing ping message")
		if err := conn.SetWriteDeadline(time.Now().Add(h.Config.WriteWait)); err != nil {
			h.logger.WithField("error", err.Error()).Error("error while setting write deadline")
			return
		}

		if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
			h.logger.WithField("error", err.Error()).Error("error while sending ping message")
			return
		}
		h.logger.WithField("username", un).Debug("ping sent")
	}
}

// writer launches channel listeners in background which will receive messages from topics user is subscribed to.
func (h *SockHub) writer(un string, conn *websocket.Conn, subs []*hub.Subscription) {
	for _, s := range subs {
		// pass redis messages to user
		go func(s *hub.Subscription) {
			h.logger.WithField("topic", s.Topic).Debug("listening to message channel")
			for msg := range s.MessageChannel {
				h.logger.
					WithField("channel", msg.Channel).
					WithField("payload", msg.Payload).
					Info("message received from redis")

				err := conn.WriteMessage(1, []byte(msg.Payload))
				if err != nil {
					h.logger.WithField("error", err).Error("error while sending message to user")
					break
				}
			}
			h.logger.WithField("topic", s.Topic).Debug("message channel closed")
		}(s)
		h.logger.
			WithField("username", un).
			WithField("topics", s.Topic).
			Info("message channel listeners created")
	}
}

// reader reads user messages and then publishes them to specified topic.
func (h *SockHub) reader(ctx context.Context, username string, conn *websocket.Conn) {
	// read user sent messages
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		cm := &ClientMessage{}
		if err = json.Unmarshal(message, cm); err != nil {
			h.logger.
				WithField("username", username).
				WithField("type", mt).
				WithField("payload", cm).
				Info("invalid error from usre")
			continue
		}
		h.logger.
			WithField("username", username).
			WithField("type", mt).
			WithField("payload", cm).
			Info("message received from user")
		// TODO authorize user access to the topic.
		h.Hub.Publish(ctx, cm.Topic, cm.Body)
	}
}
