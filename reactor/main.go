package main

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/chitoku-k/ejaculation-counter/reactor/application/server"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/hardcoding"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/invoker"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/queue"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
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

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: env.LogLevel})
	slog.SetDefault(slog.New(handler))

	db, err := client.NewDB(env)
	if err != nil {
		slog.Error("Failed to initialize DB", slog.Any("err", err))
		os.Exit(1)
	}

	reader, err := queue.NewReader("ejaculation-counter.packets", "ejaculation-counter.packets.queue", "packets", env)
	if err != nil {
		slog.Error("Failed to initialize reader", slog.Any("err", err))
		os.Exit(1)
	}

	wg.Add(1)
	go func() {
		reader.Consume(ctx)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		<-ctx.Done()
		err := reader.Close(true)
		if err != nil {
			slog.Error("Failed to close reader", slog.Any("err", err))
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		c, err := client.NewHttpClient()
		if err != nil {
			slog.Error("Failed to initialize Cookie Jar", slog.Any("err", err))
			os.Exit(1)
		}
		shindan := client.NewShindanmaker(c)
		through := hardcoding.NewThroughRepository()
		doublet := hardcoding.NewDoubletRepository()
		mpyw := client.NewMpyw(c)

		mc := client.NewMastodon(env)
		ps := service.NewProcessor(
			reader,
			invoker.NewReply(mc),
			invoker.NewIncrement(env, mc, db),
			invoker.NewUpdate(env, mc, db),
			invoker.NewAdministration(mc, db),
			[]service.Action{
				action.NewOfufutonChallenge(rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())), env),
				action.NewDB(env),
				action.NewPyuUpdate(env),
				action.NewMpyw(mpyw, env),
				action.NewAVShindanmaker(shindan, env),
				action.NewBattleChimpoShindanmaker(shindan, env),
				action.NewBlueArchiveEcchiGameShindanmaker(shindan, env),
				action.NewChimpoChallengeShindanmaker(shindan, env),
				action.NewChimpoInsertionChallengeShindanmaker(shindan, env),
				action.NewChimpoMatchingShindanmaker(shindan, env),
				action.NewLawChallengeShindanmaker(shindan, env),
				action.NewOfutonManagerShindanmaker(shindan, env),
				action.NewPyuppyuManagerShindanmaker(shindan, env),
				action.NewSushiShindanmaker(shindan, env),
				action.NewThrough(through, env),
				action.NewDoublet(doublet, env),
			},
			time.Now,
		)
		ps.Execute(ctx, reader.Packets())

		err = db.Close()
		if err != nil {
			slog.Error("Failed to close connection to DB", slog.Any("err", err))
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		through := service.NewThrough(hardcoding.NewThroughRepository())
		doublet := service.NewDoublet(hardcoding.NewDoubletRepository())
		engine := server.NewEngine(env.Port, env.TLSCert, env.TLSKey, through, doublet)
		err = engine.Start(ctx)
		if err != nil {
			slog.Error("Failed to start web server", slog.Any("err", err))
			os.Exit(1)
		}
		wg.Done()
	}()

	wg.Wait()
}
