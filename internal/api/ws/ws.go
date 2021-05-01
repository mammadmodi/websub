package ws

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mammadmodi/webis/internal/source"
	"github.com/sirupsen/logrus"
	"html/template"
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
	username := r.URL.Query().Get("name")
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

func Home(ctx *gin.Context) {
	name := ctx.Request.URL.Query().Get("name")
	if err := homeTemplate.Execute(ctx.Writer, "ws://"+"127.0.0.1:8379/v1/socket/connect?name="+name); err != nil {
		ctx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx.Writer.WriteHeader(http.StatusOK)
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
