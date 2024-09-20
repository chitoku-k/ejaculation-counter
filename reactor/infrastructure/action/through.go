package action

import (
	"context"
	"io"
	"math/rand/v2"
	"regexp"
	"strconv"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/reader"
	"github.com/chitoku-k/ejaculation-counter/reactor/repository"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	ThroughRegex = regexp.MustCompile(`(?:\s*(\d+)\s*連)?駿河茶|今日の\s*through|through\s*(?:が|ガ|ｶﾞ)[チﾁ][ャｬ]`)
)

type through struct {
	Repository  repository.ThroughRepository
	Environment config.Environment
}

func NewThrough(repository repository.ThroughRepository, environment config.Environment) service.Action {
	return &through{
		Repository:  repository,
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

	if index == nil || matches == nil {
		return nil, 0, service.ErrNoMatch
	}

	count, err := strconv.Atoi(matches[1])
	if err != nil {
		count = 1
	}

	items := t.Repository.Get()
	generator := func() string {
		return items[rand.IntN(len(items))]
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        io.NopCloser(reader.NewStringFuncReader("\n", count, generator)),
		Visibility:  message.Visibility,
	}

	return event, index[0], nil
}
