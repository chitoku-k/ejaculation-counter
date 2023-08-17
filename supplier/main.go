package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"

	"github.com/chitoku-k/ejaculation-counter/supplier/application/server"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/queue"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/scheduler"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/streaming"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/sirupsen/logrus"
)

var signals = []os.Signal{os.Interrupt}

func init() {
	prometheus.DefaultRegisterer.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	prometheus.DefaultRegisterer.Unregister(collectors.NewGoCollector())
}

func main() {
	var wg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), signals...)
	defer stop()

	env, err := config.Get()
	if err != nil {
		logrus.Fatalf("Failed to initialize config: %v", err)
	}
	logrus.SetLevel(env.LogLevel)

	s, err := scheduler.New(env)
	if err != nil {
		logrus.Fatalf("Failed to initialize scheduler: %v", err)
	}
	tick := s.Start()

	wg.Add(1)
	go func() {
		<-ctx.Done()
		s.Stop()
		wg.Done()
	}()

	writer, err := queue.NewWriter(ctx, "ejaculation-counter.packets", "packets", env)
	if err != nil {
		logrus.Fatalf("Failed to initialize writer: %v", err)
	}

	mastodon := streaming.NewMastodon(
		env,
		wrapper.NewDialer(websocket.DefaultDialer),
		wrapper.NewTimer(),
	)

	wg.Add(1)
	go func() {
		err := mastodon.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			logrus.Fatalf("Error in starting streaming: %v", err)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		<-ctx.Done()
		mastodon.Close(true)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		ps := service.NewProcessor(writer)
		ps.Execute(ctx, tick, mastodon.Statuses())

		err := writer.Close()
		if err != nil {
			logrus.Errorf("Failed to close writer: %v", err)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		engine := server.NewEngine(env.Port, env.TLSCert, env.TLSKey)
		err := engine.Start(ctx)
		if err != nil {
			logrus.Fatalf("Failed to start web server: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()
}
