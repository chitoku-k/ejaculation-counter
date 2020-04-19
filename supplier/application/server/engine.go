package server

import (
	"context"
	"net/http"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type engine struct {
	ctx         context.Context
	Environment config.Environment
}

type Engine interface {
	Start() error
}

func NewEngine(ctx context.Context, environment config.Environment) Engine {
	return &engine{
		ctx:         ctx,
		Environment: environment,
	}
}

func (e *engine) Start() error {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/healthz", "/metrics"},
	}))

	router.Any("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	server := http.Server{
		Addr:    ":" + e.Environment.Port,
		Handler: router,
	}

	go func() {
		<-e.ctx.Done()
		server.Shutdown(context.Background())
	}()

	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}
