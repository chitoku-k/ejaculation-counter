package main

import (
	"context"
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
	"github.com/pkg/errors"
)

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
		panic(errors.Wrap(err, "failed to initialize config"))
	}

	reader, err := queue.NewReader("events_topic", "events", "events", env)
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize reader"))
	}

	db, err := client.NewDB(env)
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize DB"))
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
		panic(errors.Wrap(err, "failed to execute processor"))
	}

	through := service.NewThrough(hardcoding.NewThroughRepository())
	engine := server.NewEngine(env, through)
	err = engine.Start(ctx)
	if err != nil {
		panic(errors.Wrap(err, "failed to start web server"))
	}
}
