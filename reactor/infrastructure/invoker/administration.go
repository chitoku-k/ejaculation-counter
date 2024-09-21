package invoker

import (
	"context"
	"fmt"
	"strconv"
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

func (a *administration) format(acct string, result []string, affected int64) string {
	n := len("@") + len(acct) + len("\n")
	for _, r := range result {
		n += len(r) + len(separator)
	}
	n += len("(100000 rows)")

	var sb strings.Builder
	sb.Grow(n)

	sb.WriteString("@")
	sb.WriteString(acct)
	sb.WriteString("\n")

	for _, r := range result {
		sb.WriteString(r)
		sb.WriteString(separator)
	}

	sb.WriteString("(")
	sb.WriteString(strconv.FormatInt(affected, 10))

	switch affected {
	case 1:
		sb.WriteString(" row)")
	default:
		sb.WriteString(" rows)")
	}

	return sb.String()
}

func (a *administration) Do(ctx context.Context, event service.AdministrationEvent) error {
	if event.Type != "DB" {
		ExecutedAdministrationEventsErrorsTotal.Inc()
		return fmt.Errorf("failed to handle event type: %s", event.Type)
	}

	result, affected, err := a.DB.Query(ctx, event.Command)
	if err != nil {
		ExecutedAdministrationEventsErrorsTotal.Inc()
		return fmt.Errorf("failed to run query: %w", err)
	}

	status, n, err := pack(strings.NewReader(a.format(event.Acct, result, affected)))
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
