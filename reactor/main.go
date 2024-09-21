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
	slog.SetLogLoggerLevel(env.LogLevel)

	db, err := client.NewDB(map[string]string{
		"user":        env.DB.Username,
		"password":    env.DB.Password,
		"host":        env.DB.Host,
		"dbname":      env.DB.Database,
		"sslmode":     env.DB.SSLMode,
		"sslcert":     env.DB.SSLCert,
		"sslkey":      env.DB.SSLKey,
		"sslrootcert": env.DB.SSLRootCert,
	}, env.DB.MaxLifetime)
	if err != nil {
		slog.Error("Failed to initialize DB", slog.Any("err", err))
		os.Exit(1)
	}

	reader, err := queue.NewReader(
		"ejaculation-counter.packets", "ejaculation-counter.packets.queue", "packets",
		env.Queue.Host, env.Queue.Username, env.Queue.Password,
		env.Queue.SSLCert, env.Queue.SSLKey, env.Queue.SSLRootCert,
	)
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

		mc := client.NewMastodon(env.Mastodon.ServerURL, env.Mastodon.AccessToken)
		ps := service.NewProcessor(
			reader,
			invoker.NewReply(mc),
			invoker.NewIncrement(mc, db, env.UserID),
			invoker.NewUpdate(mc, db, env.UserID),
			invoker.NewAdministration(mc, db),
			[]service.Action{
				action.NewOfufutonChallenge(rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())), env.Mastodon.UserID),
				action.NewDB(env.Mastodon.UserID),
				action.NewPyuUpdate(env.Mastodon.UserID),
				action.NewMpyw(mpyw, env.Mastodon.UserID, env.External.MpywAPIURL),
				action.NewAVShindanmaker(shindan, env.Mastodon.UserID),
				action.NewBattleChimpoShindanmaker(shindan, env.Mastodon.UserID),
				action.NewBlueArchiveEcchiGameShindanmaker(shindan, env.Mastodon.UserID),
				action.NewChimpoChallengeShindanmaker(shindan, env.Mastodon.UserID),
				action.NewChimpoInsertionChallengeShindanmaker(shindan, env.Mastodon.UserID),
				action.NewChimpoMatchingShindanmaker(shindan, env.Mastodon.UserID),
				action.NewLawChallengeShindanmaker(shindan, env.Mastodon.UserID),
				action.NewOfutonManagerShindanmaker(shindan, env.Mastodon.UserID),
				action.NewPyuppyuManagerShindanmaker(shindan, env.Mastodon.UserID),
				action.NewSushiShindanmaker(shindan, env.Mastodon.UserID),
				action.NewThrough(through, env.Mastodon.UserID),
				action.NewDoublet(doublet, env.Mastodon.UserID),
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
		engine := server.NewEngine(through, doublet, env.Port, env.TLSCert, env.TLSKey)
		err := engine.Start(ctx)
		if err != nil {
			slog.Error("Failed to start web server", slog.Any("err", err))
			os.Exit(1)
		}
		wg.Done()
	}()

	wg.Wait()
}
