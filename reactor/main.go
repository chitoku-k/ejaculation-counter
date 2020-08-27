package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chitoku-k/ejaculation-counter/reactor/application/server"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/hardcoding"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/queue"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.DefaultRegisterer.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.DefaultRegisterer.Unregister(prometheus.NewGoCollector())
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer close(sig)
		<-sig
		cancel()
	}()

	env, err := config.Get()
	if err != nil {
		panic(fmt.Errorf("failed to initialize config: %w", err))
	}

	reader, err := queue.NewReader("events_topic", "events", "events", env)
	if err != nil {
		panic(fmt.Errorf("failed to initialize reader: %w", err))
	}

	db, err := client.NewDB(env)
	if err != nil {
		panic(fmt.Errorf("failed to initialize DB: %w", err))
	}

	mc := client.NewMastodon(env)
	ps := service.NewProcessor(
		reader,
		actions.NewReply(mc),
		actions.NewIncrement(env, mc, db),
		actions.NewUpdate(env, mc, db),
		actions.NewAdministration(mc, db),
	)
	err = ps.Execute(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to execute processor: %w", err))
	}

	through := service.NewThrough(hardcoding.NewThroughRepository())
	doublet := service.NewDoublet(hardcoding.NewDoubletRepository())
	engine := server.NewEngine(env, through, doublet)
	err = engine.Start(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to start web server: %w", err))
	}
}
