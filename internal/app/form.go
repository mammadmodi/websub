package app

import (
	"fmt"
	"html/template"
	"net/http"
)

// Home is a http handler that renders a html form that can be create web socket
// connection with the websocket server.
func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	topics := r.URL.Query().Get("topics")
	urlFormat := "ws://%s:%d/socket/connect?username=%s&topics=%s"
	socketUrl := fmt.Sprintf(urlFormat, a.Config.Addr, a.Config.Port, username, topics)
	if err := homeTemplate.Execute(w, socketUrl); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
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
    var topic = document.getElementById("topic");
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
            print("Message: " + evt.data);
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
        print("SEND: " + input.value + " to topic " + topic.value);
	    var obj = new Object();
		obj.body = input.value
		obj.topic = topic.value
		var message = JSON.stringify(obj);
        ws.send(message);
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
<p><input id="topic" type="text" value="Topic">
<button id="send">Send Message</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
