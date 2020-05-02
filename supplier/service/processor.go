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
	ctx       context.Context
	Scheduler Scheduler
	Streaming Streaming
	Queue     Queue
	Actions   []Action
}

type Processor interface {
	Execute()
}

func NewProcessor(
	ctx context.Context,
	scheduler Scheduler,
	streaming Streaming,
	queue Queue,
	actions []Action,
) Processor {
	return &processor{
		ctx:       ctx,
		Scheduler: scheduler,
		Streaming: streaming,
		Queue:     queue,
		Actions:   actions,
	}
}

func (ps *processor) Execute() {
	go func() {
		stream, err := ps.Streaming.Run()
		if err != nil {
			logrus.Errorln("Error in starting streaming: " + err.Error())
			return
		}

		scheduler := ps.Scheduler.Start()

		for {
			select {
			case <-ps.ctx.Done():
				return

			case event := <-scheduler:
				EventsTotal.WithLabelValues(event.Name(), "").Inc()
				err := ps.Queue.Write(event)
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
					err := ps.Queue.Write(r.Event)
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
