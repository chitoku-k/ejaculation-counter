package server

import (
	"context"
	"net/http"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type engine struct {
	Environment config.Environment
	Through     service.Through
}

type Engine interface {
	Start(ctx context.Context) error
}

func NewEngine(
	environment config.Environment,
	through service.Through,
) Engine {
	return &engine{
		Environment: environment,
		Through:     through,
	}
}

func (e *engine) Start(ctx context.Context) error {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/healthz", "/metrics"},
	}))

	router.Any("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/through", e.HandleThrough)

	server := http.Server{
		Addr:    ":" + e.Environment.Port,
		Handler: router,
	}

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}
