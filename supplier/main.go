package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chitoku-k/ejaculation-counter/supplier/application/server"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/action"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/queue"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/scheduler"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/streaming"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/gorilla/websocket"
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

	writer, err := queue.NewWriter(ctx, "events_topic", "events", env)
	if err != nil {
		panic(fmt.Errorf("failed to initialize writer: %w", err))
	}

	rand.Seed(time.Now().Unix())

	shindan := client.NewShindanmaker(http.DefaultClient)
	through := client.NewThrough(http.DefaultClient)
	doublet := client.NewDoublet(http.DefaultClient)
	mpyw := client.NewMpyw(http.DefaultClient)

	s, err := scheduler.New(env)
	if err != nil {
		panic(fmt.Errorf("failed to initialize scheduler: %w", err))
	}

	mastodon := streaming.NewMastodon(
		env,
		wrapper.NewDialer(websocket.DefaultDialer),
		wrapper.NewTimer(),
	)
	ps := service.NewProcessor(s, mastodon, writer, []service.Action{
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
		action.NewThrough(through, env),
		action.NewDoublet(doublet, env),
	})
	ps.Execute(ctx)

	engine := server.NewEngine(env)
	err = engine.Start(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to start web server: %w", err))
	}
}
