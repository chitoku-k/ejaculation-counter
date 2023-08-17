package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"sync"

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

	db, err := client.NewDB(env)
	if err != nil {
		logrus.Fatalf("Failed to initialize DB: %v", err)
	}

	reader, err := queue.NewReader("packets_topic", "packets", "packets", env)
	if err != nil {
		logrus.Fatalf("Failed to initialize reader: %v", err)
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
			logrus.Errorf("Failed to close reader: %v", err)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		c, err := client.NewHttpClient()
		if err != nil {
			logrus.Fatalf("Failed to initialize Cookie Jar: %v", err)
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
				action.NewOfufutonChallenge(rand.New(rand.NewSource(1)), env),
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
		)
		ps.Execute(ctx, reader.Packets())

		err = db.Close()
		if err != nil {
			logrus.Errorf("Failed to close connection to DB: %v", err)
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
			logrus.Fatalf("Failed to start web server: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()
}
