package main

import (
	"github.com/go-redis/redis"
	"os"
	"webchannel/ws"
)

func main() {
	os.Exit(run())
}

func run() int {
	es := newEventSource()
	if err := ws.ServeWS(es); err != nil {
		return 1
	}
	return 0
}

func newEventSource() *ws.RedisEventSource {
	es := &ws.RedisEventSource{
		Client: newRedisCli("127.0.0.1:6379", ""),
	}
	return es
}

func newRedisCli(addr, pass string) *redis.Client {
	opts := &redis.Options{
		Addr:     addr,
		Password: pass,
	}
	c := redis.NewClient(opts)
	return c
}
