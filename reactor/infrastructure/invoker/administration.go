package invoker

import (
	"context"
	"fmt"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/mattn/go-mastodon"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	separator = "\n--------\n"
)

var (
	ExecutedAdministrationEventsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "executed_administration_events_total",
		Help:      "Total number of administration events replied through API.",
	})
	ExecutedAdministrationEventsErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "executed_administration_events_errors_total",
		Help:      "Total number of errors triggered when replying to administration through API.",
	})
)

type administration struct {
	Client *mastodon.Client
	DB     client.DB
}

func NewAdministration(
	client *mastodon.Client,
	db client.DB,
) service.Administration {
	return &administration{
		Client: client,
		DB:     db,
	}
}

func (a *administration) Do(ctx context.Context, event service.AdministrationEvent) error {
	if event.Type != "DB" {
		ExecutedAdministrationEventsErrorsTotal.Inc()
		return fmt.Errorf("failed to handle event type: %s", event.Type)
	}

	result, err := a.DB.Query(ctx, event.Command)
	if err != nil {
		ExecutedAdministrationEventsErrorsTotal.Inc()
		return fmt.Errorf("failed to run query: %w", err)
	}

	status, n, err := pack(strings.NewReader(fmt.Sprintf("@%s\n%s", event.Acct, strings.Join(result, separator))))
	if err != nil {
		ExecutedAdministrationEventsErrorsTotal.Inc()
		return fmt.Errorf("failed to prepare reply (%v bytes): %w", n, err)
	}

	_, err = a.Client.PostStatus(ctx, &mastodon.Toot{
		InReplyToID: mastodon.ID(event.InReplyToID),
		Status:      status,
		Visibility:  event.Visibility,
	})
	if err != nil {
		ExecutedAdministrationEventsErrorsTotal.Inc()
		return fmt.Errorf("failed to send reply: %w", err)
	}

	ExecutedAdministrationEventsTotal.Inc()
	return nil
}
