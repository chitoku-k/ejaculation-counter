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
	DoubletRegex = regexp.MustCompile(`今日の\s*(?:doublet|二重語)|(?:\s*(\d+)\s*連\s*)?(?:doublet|二重語)\s*(?:が|ガ|ｶﾞ)[チﾁ][ャｬ]`)
)

type doublet struct {
	Client      client.Doublet
	Environment config.Environment
}

func NewDoublet(c client.Doublet, environment config.Environment) service.Action {
	return &doublet{
		Client:      c,
		Environment: environment,
	}
}

func (t *doublet) Name() string {
	return "二重語ガチャ"
}

func (t *doublet) Target(message service.Message) bool {
	return !message.IsReblog &&
		(message.Account.ID != t.Environment.Mastodon.UserID || message.InReplyToID == "") &&
		DoubletRegex.MatchString(message.Content)
}

func (t *doublet) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := DoubletRegex.FindStringIndex(message.Content)
	matches := DoubletRegex.FindStringSubmatch(message.Content)

	count, err := strconv.Atoi(matches[1])
	if err != nil {
		count = 1
	}

	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort("localhost", t.Environment.Port),
		Path:   "/doublet",
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
