package service

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

const (
	PacketTTL = 30 * time.Minute
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
	Clock          func() time.Time
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
	clock func() time.Time,
) Processor {
	return &processor{
		Queue:          queue,
		Reply:          reply,
		Increment:      increment,
		Update:         update,
		Administration: administration,
		Actions:        actions,
		Clock:          clock,
	}
}

func (ps *processor) Execute(ctx context.Context, packets <-chan Packet) {
	for packet := range packets {
		if ps.Clock().Sub(packet.Timestamp()) > PacketTTL {
			logrus.WithFields(logrus.Fields{"message-timestamp": packet.Timestamp()}).Warnln("A message has been discarded as it is too old.")
			ps.Queue.Ack(packet.Tag())
			continue
		}

		switch p := packet.(type) {
		case Tick:
			go func() {
				err := ps.Update.Do(ctx, UpdateEvent{
					Year:  p.Year,
					Month: p.Month,
					Day:   p.Day,
				})
				if err != nil {
					logrus.Errorf("Failed to update: %v", err)
					ps.Queue.Reject(p.Tag())
					return
				}
				ps.Queue.Ack(p.Tag())
			}()

		case Message:
			go func() {
				var result []actionResult
				for _, action := range ps.Actions {
					if !action.Target(p) {
						continue
					}

					event, index, err := action.Event(ctx, p)
					if err != nil {
						logrus.Errorf("Error in processing %v: %v", action.Name(), err)
						EventsErrorTotal.WithLabelValues(action.Name()).Inc()
						result = append(result, actionResult{
							Event: ReplyErrorEvent{
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

				slices.SortFunc(result, func(a, b actionResult) int {
					return cmp.Compare(a.Index, b.Index)
				})

				err := ps.doEvents(ctx, result)
				if err != nil {
					logrus.Errorf("Failed to process: %v", err)
					ps.Queue.Reject(p.Tag())
					return
				}
				ps.Queue.Ack(p.Tag())
			}()
		}
	}
}

func (ps *processor) doEvents(ctx context.Context, result []actionResult) error {
	var errs error
	for _, r := range result {
		switch event := r.Event.(type) {
		case ReplyEvent:
			err := ps.Reply.Send(ctx, event)
			if err != nil {
				errs = errors.Join(errs, err)
			}

		case ReplyErrorEvent:
			err := ps.Reply.SendError(ctx, event)
			if err != nil {
				errs = errors.Join(errs, err)
			}

		case IncrementEvent:
			err := ps.Increment.Do(ctx, event)
			if err != nil {
				errs = errors.Join(errs, err)
			}

		case AdministrationEvent:
			err := ps.Administration.Do(ctx, event)
			if err != nil {
				logrus.Errorf("Failed to execute administrative operation: %v", err)
			}

		default:
			errs = errors.Join(errs, fmt.Errorf("failed to process unknown event: %v", event.Name()))
		}
	}

	return errs
}
