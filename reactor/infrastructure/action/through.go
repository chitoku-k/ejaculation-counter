package action

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/url"
	"regexp"
	"strconv"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/reader"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	ThroughRegex = regexp.MustCompile(`(?:\s*(\d+)\s*連)?駿河茶|今日の\s*through|through\s*(?:が|ガ|ｶﾞ)[チﾁ][ャｬ]`)
)

type through struct {
	Client      client.Through
	Environment config.Environment
}

func NewThrough(c client.Through, environment config.Environment) service.Action {
	return &through{
		Client:      c,
		Environment: environment,
	}
}

func (t *through) Name() string {
	return "駿河茶"
}

func (t *through) Target(message service.Message) bool {
	return !message.IsReblog &&
		(message.Account.ID != t.Environment.Mastodon.UserID || message.InReplyToID == "") &&
		ThroughRegex.MatchString(message.Content)
}

func (t *through) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := ThroughRegex.FindStringIndex(message.Content)
	matches := ThroughRegex.FindStringSubmatch(message.Content)

	count, err := strconv.Atoi(matches[1])
	if err != nil {
		count = 1
	}

	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort("localhost", t.Environment.Port),
		Path:   "/through",
	}

	items, err := t.Client.Do(ctx, u.String())
	if err != nil {
		return nil, index[0], fmt.Errorf("failed to create event: %w", err)
	}

	generator := func() string {
		return items[rand.Intn(len(items))]
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        io.NopCloser(reader.NewStringFuncReader("\n", count, generator)),
		Visibility:  message.Visibility,
	}

	return event, index[0], nil
}
