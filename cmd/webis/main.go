/**
webis main command is used to run a ws server
*/
package main

import (
	"context"
	"fmt"
	"github.com/mammadmodi/webis/internal/api/ws"
	"github.com/mammadmodi/webis/internal/app"
	"github.com/mammadmodi/webis/internal/hub"
	"github.com/mammadmodi/webis/pkg/logger"
	"github.com/mammadmodi/webis/pkg/redis"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strings"
)

// Following variables must be loaded in build time.
var (
	CommitSHA     string
	CommitRefName string
	BuildDate     string
)

var (
	c *app.Configs
	l *logrus.Logger
	a *app.App
)

func init() {
	// initializing configuration
	var err error
	c, err = app.NewConfiguration()
	if err != nil {
		panic(fmt.Errorf("error while initializing configs, error: %v", err))
	}

	//init logger
	l, err = logger.NewLogrusLogger(c.LoggingConfigs)
	if err != nil {
		panic(fmt.Errorf("error while initializing logger, error: %v", err))
	}

	// initializing redis client
	rc, err := redis.NewClient(c.RedisConfigs)
	if err != nil {
		l.Fatalf("error while initializing redis client, error: %v", err)
	}
	_, err = rc.Ping().Result()
	if err != nil {
		l.Fatalf("cannot get ping response with redis client, error: %v", err)
	}

	rh := hub.NewRedisHub(rc, hub.RedisHubConfig{})
	sh := ws.NewSockHub(rh)

	// initializing application instance
	a = &app.App{
		Config:  c,
		Logger:  l,
		SockHub: sh,
	}
}

func main() {
	// printing ascii art
	asciiArt := strings.Replace(app.AsciiArt, "bt", "`", 2)
	fmt.Println(strings.NewReplacer("__commit_ref_name__", CommitRefName, "__commit_sha__", CommitSHA, "__build_date__", BuildDate).Replace(asciiArt))

	// listen on os signal in background
	ctx, cancel := context.WithCancel(context.Background())
	go listenOnOsSignal(cancel)

	// starting app
	if err := a.Start(ctx); err != nil {
		l.Fatalf("error while starting app, error: %s", err.Error())
	}

	// wait until main context done
	<-ctx.Done()

	// shutting down app gracefully
	a.Stop(context.Background())
	l.Exit(0)
	os.Exit(0)
}

// listenOnOsSignal listens on os signal and calls cancel func when got interrupt sig
func listenOnOsSignal(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill)
	sig := <-sigCh
	l.Warnf("got signal %s from OS, calling application's main cancel func ...", sig)
	cancel()
}
