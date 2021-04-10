package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/chitoku-k/ejaculation-counter/reactor/application/server"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/hardcoding"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/invoker"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/queue"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func init() {
	prometheus.DefaultRegisterer.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.DefaultRegisterer.Unregister(prometheus.NewGoCollector())
}

func main() {
	var wg sync.WaitGroup
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
		shindan := client.NewShindanmaker(http.DefaultClient)
		through := client.NewThrough(http.DefaultClient)
		doublet := client.NewDoublet(http.DefaultClient)
		mpyw := client.NewMpyw(http.DefaultClient)

		mc := client.NewMastodon(env)
		ps := service.NewProcessor(
			reader,
			invoker.NewReply(mc),
			invoker.NewIncrement(env, mc, db),
			invoker.NewUpdate(env, mc, db),
			invoker.NewAdministration(mc, db),
			[]service.Action{
				action.NewOfufutonChallenge(rand.New(rand.NewSource(1))),
				action.NewDB(env),
				action.NewPyuUpdate(env),
				action.NewMpyw(mpyw),
				action.NewAVShindanmaker(shindan),
				action.NewBattleChimpoShindanmaker(shindan),
				action.NewChimpoChallengeShindanmaker(shindan),
				action.NewChimpoInsertionChallengeShindanmaker(shindan),
				action.NewChimpoMatchingShindanmaker(shindan),
				action.NewLawChallengeShindanmaker(shindan),
				action.NewOfutonManagerShindanmaker(shindan),
				action.NewPyuppyuManagerShindanmaker(shindan),
				action.NewSushiShindanmaker(shindan),
				action.NewThrough(through),
				action.NewDoublet(doublet),
			},
		)
		ps.Execute(ctx, reader.Packets())

		err := db.Close()
		if err != nil {
			logrus.Errorf("Failed to close connection to DB: %v", err)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		through := service.NewThrough(hardcoding.NewThroughRepository())
		doublet := service.NewDoublet(hardcoding.NewDoubletRepository())
		engine := server.NewEngine(env.Port, through, doublet)
		err = engine.Start(ctx)
		if err != nil {
			logrus.Fatalf("Failed to start web server: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()
}
