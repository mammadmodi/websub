package app

import (
	"context"
	"fmt"
	"github.com/mammadmodi/webis/internal/api/websocket"
	//"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

const AsciiArt = `
 _    _  ____  ____  ____  ___ 
( \/\/ )( ___)(  _ \(_  _)/ __)
 )    (  )__)  ) _ < _)(_ \__ \
(__/\__)(____)(____/(____)(___/

Version: __commit_ref_name__ (__commit_sha__)
Build Date: __build_date__
`

// App is a type that serves webis functionality
type App struct {
	Config  *Configs
	Logger  *logrus.Logger
	SockHub *websocket.SockHub

	server *http.Server
}

// Start runs api server in background
func (a *App) Start(ctx context.Context) error {
	// running http server
	mux := a.initMux()
	a.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%v", a.Config.Addr, a.Config.Port),
		Handler: mux,
	}
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Panicf("error while running gin http server, error: %v", err)
		}
	}()
	return nil
}

// Stop stops the application and http server.
func (a *App) Stop(ctx context.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), a.Config.GracefulTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctxWithTimeout); err != nil {
		a.Logger.Errorf("failed to gracefully shutdown the http server, %s", err)
	} else {
		a.Logger.Info("http server closed successfully")
	}

	return
}

// initMux returns a ServeMux that routes requests to available apis.
func (a *App) initMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/socket/form", a.Home)
	mux.HandleFunc("/socket/connect", a.SockHub.Connect)

	return mux
}
