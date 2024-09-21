package main

import (
	"context"
	"errors"
	"log/slog"
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
		slog.Error("Failed to initialize config", slog.Any("err", err))
		os.Exit(1)
	}
	slog.SetLogLoggerLevel(env.LogLevel)

	s, err := scheduler.New()
	if err != nil {
		slog.Error("Failed to initialize scheduler", slog.Any("err", err))
		os.Exit(1)
	}
	tick := s.Start()

	wg.Add(1)
	go func() {
		<-ctx.Done()
		s.Stop()
		wg.Done()
	}()

	writer, err := queue.NewWriter(
		ctx,
		"ejaculation-counter.packets", "packets",
		env.Queue.Host, env.Queue.Username, env.Queue.Password,
		env.Queue.SSLCert, env.Queue.SSLKey, env.Queue.SSLRootCert,
	)
	if err != nil {
		slog.Error("Failed to initialize writer", slog.Any("err", err))
		os.Exit(1)
	}

	mastodon := streaming.NewMastodon(
		wrapper.NewDialer(websocket.DefaultDialer),
		wrapper.NewTimer(),
		env.Mastodon.ServerURL,
		env.Mastodon.AccessToken,
		env.Mastodon.Stream,
	)

	wg.Add(1)
	go func() {
		err := mastodon.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("Error in starting streaming", slog.Any("err", err))
			os.Exit(1)
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
			slog.Error("Failed to close writer", slog.Any("err", err))
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		engine := server.NewEngine(env.Port, env.TLSCert, env.TLSKey)
		err := engine.Start(ctx)
		if err != nil {
			slog.Error("Failed to start web server", slog.Any("err", err))
			os.Exit(1)
		}
		wg.Done()
	}()

	wg.Wait()
}
