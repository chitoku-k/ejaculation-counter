package invoker

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/mattn/go-mastodon"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rivo/uniseg"
)

const (
	MaxTootLength = 500
)

var (
	MentionRegexp = regexp.MustCompile(`(@[a-z0-9_]+([a-z0-9_\.-]+[a-z0-9_]+)?)@[[:word:].-]+[a-z0-9]+`)
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
	Client *mastodon.Client
}

func NewReply(client *mastodon.Client) service.Reply {
	return &reply{
		Client: client,
	}
}

func getTootLength(s string) int {
	s = MentionRegexp.ReplaceAllString(s, "$1")
	return uniseg.GraphemeClusterCount(s)
}

func pack(r io.Reader) (string, int, error) {
	buf := make([]byte, 512)

	var builder strings.Builder
	builder.Grow(1024)

	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return builder.String(), builder.Len(), err
		}
		if getTootLength(builder.String()+string(buf[:n])) > MaxTootLength {
			break
		}
		builder.Write(buf[:n])
	}

	return builder.String(), builder.Len(), nil
}

func (r *reply) Send(ctx context.Context, event service.ReplyEvent) error {
	defer func() {
		io.Copy(io.Discard, event.Body)
		event.Body.Close()
	}()

	status, n, err := pack(io.MultiReader(strings.NewReader(fmt.Sprintf("@%s ", event.Acct)), event.Body))
	if err != nil {
		RepliedEventsErrorTotal.Inc()
		return fmt.Errorf("failed to prepare reply (%v bytes): %w", n, err)
	}

	_, err = r.Client.PostStatus(ctx, &mastodon.Toot{
		InReplyToID: mastodon.ID(event.InReplyToID),
		Status:      status,
		Visibility:  event.Visibility,
	})
	if err != nil {
		RepliedEventsErrorTotal.Inc()
		return fmt.Errorf("failed to send reply: %w", err)
	}

	RepliedEventsTotal.Inc()
	return nil
}

func (r *reply) SendError(ctx context.Context, event service.ReplyErrorEvent) error {
	_, err := r.Client.PostStatus(ctx, &mastodon.Toot{
		InReplyToID: mastodon.ID(event.InReplyToID),
		Status:      fmt.Sprintf("@%s 何かがおかしいよ（%s）", event.Acct, event.ActionName),
		Visibility:  event.Visibility,
	})
	if err != nil {
		RepliedEventsErrorTotal.Inc()
		return fmt.Errorf("failed to send reply (error): %w", err)
	}

	RepliedEventsTotal.Inc()
	return nil
}
