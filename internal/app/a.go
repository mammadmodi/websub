package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mammadmodi/webis/internal/api/ws"
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
	Config    *Configs
	Logger    *logrus.Logger
	WSManager *ws.SocketManager

	server *http.Server
}

// Start runs api server in background
func (a *App) Start(ctx context.Context) error {
	// running http server
	r := a.initRouter()
	a.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%v", a.Config.Addr, a.Config.Port),
		Handler: r,
	}
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Panicf("error while running gin http server, error: %v", err)
		}
	}()
	return nil
}

// Stop stops handler's worker pool and gin http server
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

// initRouter initializes a router for ack endpoints by application's config
func (a *App) initRouter() *gin.Engine {
	gin.SetMode(a.Config.Mode)

	// init gin router
	r := gin.New()
	if a.Config.Mode == gin.DebugMode {
		r.Use(gin.Logger())
	}
	//r.GET("/healthz", a.Handler.Health)
	//r.GET("/metrics", a.Handler.MetricsMiddleware, gin.WrapH(promhttp.Handler()))
	//r.GET("/v1/socket", a.Handler.ResolveAckMiddleware, a.Handler.Ack)

	r.GET("/v1/socket/form", a.Home)
	r.GET("/v1/socket/connect", a.WSManager.Socket)

	return r
}
