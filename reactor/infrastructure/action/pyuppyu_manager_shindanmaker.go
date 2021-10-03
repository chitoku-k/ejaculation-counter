package action

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	PyuppyuManagerRegex = regexp.MustCompile(`ぴゅっぴゅしても?[いよ良]い[?？]`)
)

type pyuppyuManagerShindanmaker struct {
	Client      client.Shindanmaker
	Environment config.Environment
}

func NewPyuppyuManagerShindanmaker(c client.Shindanmaker, environment config.Environment) service.Action {
	return &pyuppyuManagerShindanmaker{
		Client:      c,
		Environment: environment,
	}
}

func (ps *pyuppyuManagerShindanmaker) Name() string {
	return "おちんちんぴゅっぴゅ管理官の毎日"
}

func (ps *pyuppyuManagerShindanmaker) Target(message service.Message) bool {
	return !message.IsReblog &&
		(message.Account.ID != ps.Environment.Mastodon.UserID || message.InReplyToID == "") &&
		PyuppyuManagerRegex.MatchString(message.Content)
}

func (ps *pyuppyuManagerShindanmaker) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := PyuppyuManagerRegex.FindStringIndex(message.Content)
	result, err := ps.Client.Do(ctx, ps.Client.Name(message.Account), "https://shindanmaker.com/a/503598")
	if err != nil {
		return nil, index[0], fmt.Errorf("failed to create event: %w", err)
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        io.NopCloser(strings.NewReader(result)),
		Visibility:  message.Visibility,
	}

	return event, index[0], nil
}
