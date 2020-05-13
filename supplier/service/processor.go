package service

import (
	"context"
	"sort"

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
		stream, err := ps.Streaming.Run(ctx)
		if err != nil {
			logrus.Errorln("Error in starting streaming: " + err.Error())
			return
		}

		scheduler := ps.Scheduler.Start()

		for {
			select {
			case <-ctx.Done():
				return

			case event := <-scheduler:
				EventsTotal.WithLabelValues(event.Name(), "").Inc()
				err := ps.Writer.Publish(event)
				if err != nil {
					logrus.Errorln("Error in queueing: " + err.Error())
					continue
				}

			case status := <-stream:
				if status.Error != nil {
					logrus.Errorln("Error in streaming: " + status.Error.Error())
					continue
				}

				_, ok := replies[status.Message.InReplyToID]
				if ok {
					continue
				}

				var result []actionResult
				for _, action := range ps.Actions {
					if !action.Target(status.Message) {
						continue
					}

					event, index, err := action.Event(status.Message)
					if err != nil {
						logrus.Errorln("Error in processing " + action.Name() + ": " + err.Error())
						EventsErrorTotal.WithLabelValues(action.Name()).Inc()
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
						logrus.Errorln("Error in queueing: " + err.Error())
						continue
					}

					replies[status.Message.ID]++
				}
			}
		}
	}()
}
