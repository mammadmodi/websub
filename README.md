## Webis

Easily make a real time communication core between your users and your microservices through web socket with power of redis.

![alt text](https://github.com/mammadmodi/webis/blob/master/architecture.png?raw=true)

## Quick Setup

Simply run ```docker-compose -f ./test/docker-compose.yaml up```.
It will set up a redis instance and a webis instance that listens to port 8379.

### Connect to Webis
In browser go to the [webis form](http://127.0.0.1:8379/socket/form?username=john&topics=johntopic1,johntopic2) and
start a websocket connection by clicking on "open" button.
It will create a websocket connection to webis and webis will create redis subscriptions to topics "johntopic1" and "johntopic2"
for the user.

### Send Messages to User
Now you can send messages to your client through redis:

#### Publish to the topic "jonhtopic1"

```docker-compose exec -f ./test/docker-compose.yaml redis /bin/sh -c "redis-cli publish johntopic1 hello-john"```

#### Publish to the topic "jonhtopic1"

```docker-compose exec -f ./test/docker-compose.yaml redis /bin/sh -c "redis-cli publish johntopic2 hello-john"```
