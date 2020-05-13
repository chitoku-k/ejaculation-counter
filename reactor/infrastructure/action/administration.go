package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/mattn/go-mastodon"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
		return errors.New("failed to handle event type: " + event.Type)
	}

	result, err := a.DB.Query(ctx, event.Command)
	if err != nil {
		ExecutedAdministrationEventsErrorsTotal.Inc()
		return errors.Wrap(err, "failed to run query")
	}

	_, err = a.Client.PostStatus(ctx, &mastodon.Toot{
		InReplyToID: mastodon.ID(event.InReplyToID),
		Status:      pack(fmt.Sprintf("@%s %s", event.Acct, strings.Join(result, "\n"))),
		Visibility:  "private",
	})
	if err != nil {
		ExecutedAdministrationEventsErrorsTotal.Inc()
		return errors.Wrap(err, "failed to send reply")
	}

	ExecutedAdministrationEventsTotal.Inc()
	return nil
}
