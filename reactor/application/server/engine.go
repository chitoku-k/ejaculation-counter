package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

type engine struct {
	Port     string
	CertFile string
	KeyFile  string
	Through  service.Through
	Doublet  service.Doublet
}

type Engine interface {
	Start(ctx context.Context) error
}

func NewEngine(
	port string,
	certFile string,
	keyFile string,
	through service.Through,
	doublet service.Doublet,
) Engine {
	return &engine{
		Port:     port,
		CertFile: certFile,
		KeyFile:  keyFile,
		Through:  through,
		Doublet:  doublet,
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
	router.GET("/doublet", e.HandleDoublet)

	server := http.Server{
		Addr:    net.JoinHostPort("", e.Port),
		Handler: router,
	}

	var eg errgroup.Group
	eg.Go(func() error {
		<-ctx.Done()
		return server.Shutdown(context.Background())
	})

	var err error
	if e.CertFile != "" && e.KeyFile != "" {
		server.TLSConfig = &tls.Config{
			GetCertificate: e.getCertificate,
		}
		err = server.ListenAndServeTLS("", "")
	} else {
		err = server.ListenAndServe()
	}

	if err == http.ErrServerClosed {
		return eg.Wait()
	}

	return err
}

func (e *engine) getCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(e.CertFile, e.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	return &cert, nil
}
