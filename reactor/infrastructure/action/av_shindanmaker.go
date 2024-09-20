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
	AVRegex = regexp.MustCompile(`([\p{L}\p{N}\p{Zs}]+?)\s*(?:くん|ちゃん)?の\s*(AV)\b`)
)

type avShindanmaker struct {
	Client      client.Shindanmaker
	Environment config.Environment
}

func NewAVShindanmaker(c client.Shindanmaker, environment config.Environment) service.Action {
	return &avShindanmaker{
		Client:      c,
		Environment: environment,
	}
}

func (as *avShindanmaker) Name() string {
	return "同人AVタイトルジェネレーター"
}

func (as *avShindanmaker) Target(message service.Message) bool {
	if message.IsReblog || (message.Account.ID == as.Environment.Mastodon.UserID && message.InReplyToID != "") {
		return false
	}

	for _, tag := range message.Tags {
		if tag.Name == "同人avタイトルジェネレーター" {
			return false
		}
	}

	return AVRegex.MatchString(message.Content)
}

func (as *avShindanmaker) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := AVRegex.FindStringSubmatchIndex(message.Content)
	matches := AVRegex.FindStringSubmatch(message.Content)

	if index == nil || matches == nil {
		return nil, 0, service.ErrNoMatch
	}

	result, err := as.Client.Do(ctx, matches[1], "https://shindanmaker.com/a/794363")
	if err != nil {
		return nil, index[4], fmt.Errorf("failed to create event: %w", err)
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        io.NopCloser(strings.NewReader(result)),
		Visibility:  message.Visibility,
	}

	return event, index[4], nil
}
