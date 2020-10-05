package service

import (
	"context"
	"sort"

	"github.com/onsi/ginkgo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

var (
	replies     = map[string]int{}
	EventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "events_total",
		Help:      "Total number of events triggered by messages.",
	}, []string{"name", "action"})
	EventsErrorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "events_error_total",
		Help:      "Total number of errors when creating events.",
	}, []string{"action"})
)

type processor struct {
	Scheduler Scheduler
	Streaming Streaming
	Writer    QueueWriter
	Actions   []Action
}

type Processor interface {
	Execute(ctx context.Context)
}

func NewProcessor(
	scheduler Scheduler,
	streaming Streaming,
	writer QueueWriter,
	actions []Action,
) Processor {
	return &processor{
		Scheduler: scheduler,
		Streaming: streaming,
		Writer:    writer,
		Actions:   actions,
	}
}

func (ps *processor) Execute(ctx context.Context) {
	go func() {
		defer ginkgo.GinkgoRecover()

		stream, err := ps.Streaming.Run(ctx)
		if err != nil {
			logrus.Errorf("Error in starting streaming: %v", err)
			return
		}

		scheduler := ps.Scheduler.Start()

		for {
			select {
			case <-ctx.Done():
				return

			case event, ok := <-scheduler:
				if !ok {
					continue
				}

				EventsTotal.WithLabelValues(event.Name(), "").Inc()
				err := ps.Writer.Publish(event)
				if err != nil {
					logrus.Errorf("Error in queueing: %v", err)
					continue
				}

			case status, ok := <-stream:
				if !ok {
					continue
				}

				switch status := status.(type) {
				case Error:
					logrus.Errorf("Error in streaming: %v", status.Err)
					continue

				case Reconnection:
					logrus.Infof("Reconnecting to streaming in %v...", status.In)
					continue

				case Message:
					_, ok := replies[status.InReplyToID]
					if ok {
						continue
					}

					var result []actionResult
					for _, action := range ps.Actions {
						if !action.Target(status) {
							continue
						}

						event, index, err := action.Event(status)
						if err != nil {
							logrus.Errorf("Error in processing %v: %v", action.Name(), err)
							EventsErrorTotal.WithLabelValues(action.Name()).Inc()
							result = append(result, actionResult{
								Event: &ReplyErrorEvent{
									InReplyToID: status.ID,
									Acct:        status.Account.Acct,
									Visibility:  status.Visibility,
									ActionName:  action.Name(),
								},
							})
							continue
						}

						result = append(result, actionResult{event, index})
						EventsTotal.WithLabelValues(event.Name(), action.Name()).Inc()
					}

					sort.Slice(result, func(i, j int) bool {
						return result[i].Index < result[j].Index
					})

					for _, r := range result {
						err := ps.Writer.Publish(r.Event)
						if err != nil {
							logrus.Errorf("Error in publishing: %v", err)
							continue
						}

						replies[status.ID]++
					}
				}
			}
		}
	}()
}
