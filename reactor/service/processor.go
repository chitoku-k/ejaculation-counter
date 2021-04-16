package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

var (
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
	Queue          QueueReader
	Reply          Reply
	Increment      Increment
	Update         Update
	Administration Administration
	Actions        []Action
}

type Processor interface {
	Execute(ctx context.Context, packets <-chan Packet)
}

func NewProcessor(
	queue QueueReader,
	reply Reply,
	increment Increment,
	update Update,
	administration Administration,
	actions []Action,
) Processor {
	return &processor{
		Queue:          queue,
		Reply:          reply,
		Increment:      increment,
		Update:         update,
		Administration: administration,
		Actions:        actions,
	}
}

func (ps *processor) Execute(ctx context.Context, packets <-chan Packet) {
	for packet := range packets {
		switch p := packet.(type) {
		case *Tick:
			go func() {
				err := ps.Update.Do(ctx, UpdateEvent{
					Year:  p.Year,
					Month: p.Month,
					Day:   p.Day,
				})
				if err != nil {
					logrus.Errorf("Failed to update: %v", err)
					return
				}
				p.Ack()
			}()

		case *Message:
			go func() {
				var result []actionResult
				for _, action := range ps.Actions {
					if !action.Target(*p) {
						continue
					}

					event, index, err := action.Event(*p)
					if err != nil {
						logrus.Errorf("Error in processing %v: %v", action.Name(), err)
						EventsErrorTotal.WithLabelValues(action.Name()).Inc()
						result = append(result, actionResult{
							Event: &ReplyErrorEvent{
								InReplyToID: p.ID,
								Acct:        p.Account.Acct,
								Visibility:  p.Visibility,
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

				err := ps.doEvents(ctx, result)
				if err == nil {
					p.Ack()
				}
			}()
		}
	}
}

func (ps *processor) doEvents(ctx context.Context, result []actionResult) error {
	var requeue bool
	for _, r := range result {
		switch event := r.Event.(type) {
		case *ReplyEvent:
			err := ps.Reply.Send(ctx, *event)
			if err != nil {
				requeue = true
				logrus.Errorf("Failed to send reply: %v", err)
			}

		case *ReplyErrorEvent:
			err := ps.Reply.SendError(ctx, *event)
			if err != nil {
				requeue = true
				logrus.Errorf("Failed to send reply: %v", err)
			}

		case *IncrementEvent:
			err := ps.Increment.Do(ctx, *event)
			if err != nil {
				requeue = true
				logrus.Errorf("Failed to increment: %v", err)
			}

		case *AdministrationEvent:
			err := ps.Administration.Do(ctx, *event)
			if err != nil {
				logrus.Errorf("Failed to execute administrative operation: %v", err)
			}

		default:
			requeue = true
			logrus.Warnf("Failed to process unknown event: %v", event.Name())
		}
	}

	if requeue {
		return fmt.Errorf("failed to process event(s)")
	}
	return nil
}
