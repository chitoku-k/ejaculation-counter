package service

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
			slog.Warn("A message has been discarded as it is too old", slog.Any("message-timestamp", packet.Timestamp()))
			_ = ps.Queue.Ack(packet.Tag())
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
					slog.Error("Failed to update", slog.Any("err", err))

					err := ps.Queue.Reject(p.Tag())
					if err != nil {
						slog.Warn("The message could not be rejected", slog.Any("err", err))
					}
					return
				}

				err = ps.Queue.Ack(p.Tag())
				if err != nil {
					slog.Warn("The message could not be acknowledged", slog.Any("err", err))
				}
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
						slog.Error("Error in processing", slog.String("action", action.Name()), slog.Any("err", err))
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
					slog.Error("Failed to process", slog.Any("err", err))

					err := ps.Queue.Reject(p.Tag())
					if err != nil {
						slog.Warn("The message could not be rejected", slog.Any("err", err))
					}
					return
				}

				err = ps.Queue.Ack(p.Tag())
				if err != nil {
					slog.Warn("The message could not be acknowledged", slog.Any("err", err))
				}
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
				slog.Error("Failed to execute administrative operation", slog.Any("err", err))
			}

		default:
			errs = errors.Join(errs, fmt.Errorf("failed to process unknown event: %v", event.Name()))
		}
	}

	return errs
}
