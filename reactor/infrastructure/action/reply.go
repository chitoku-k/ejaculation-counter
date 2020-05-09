package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/mattn/go-mastodon"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	MaxTootLength = 500
)

var (
	RepliedEventsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "replied_events_total",
		Help:      "Total number of events replied through API.",
	})
	RepliedEventsErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "replied_events_error_total",
		Help:      "Total number of errors triggered when replying through API.",
	})
)

type reply struct {
	ctx    context.Context
	Client *mastodon.Client
}

func NewReply(ctx context.Context, client *mastodon.Client) service.Reply {
	return &reply{
		ctx:    ctx,
		Client: client,
	}
}

func pack(s string) string {
	var builder strings.Builder

	for _, r := range []rune(s) {
		if builder.Len() >= MaxTootLength {
			break
		}
		builder.WriteRune(r)
	}

	return builder.String()
}

func (r *reply) Send(event service.ReplyEvent) error {
	_, err := r.Client.PostStatus(r.ctx, &mastodon.Toot{
		InReplyToID: mastodon.ID(event.InReplyToID),
		Status:      pack(fmt.Sprintf("@%s %s", event.Acct, event.Body)),
		Visibility:  event.Visibility,
	})
	if err != nil {
		RepliedEventsErrorTotal.Inc()
		return errors.Wrap(err, "failed to send reply")
	}

	RepliedEventsTotal.Inc()
	return nil
}
