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
	ChimpoInsertionChallengeRegex = regexp.MustCompile(`ちん(ちん|ぽ|こ)挿入[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)`)
)

type chimpoInsertionChallengeShindanmaker struct {
	Client      client.Shindanmaker
	Environment config.Environment
}

func NewChimpoInsertionChallengeShindanmaker(c client.Shindanmaker, environment config.Environment) service.Action {
	return &chimpoInsertionChallengeShindanmaker{
		Client:      c,
		Environment: environment,
	}
}

func (cs *chimpoInsertionChallengeShindanmaker) Name() string {
	return "おちんぽ挿入チャレンジ"
}

func (cs *chimpoInsertionChallengeShindanmaker) Target(message service.Message) bool {
	if message.IsReblog || (message.Account.ID == cs.Environment.Mastodon.UserID && message.InReplyToID != "") {
		return false
	}

	for _, tag := range message.Tags {
		if tag.Name == "おちんぽ挿入チャレンジ" {
			return false
		}
	}

	return ChimpoInsertionChallengeRegex.MatchString(message.Content)
}

func (cs *chimpoInsertionChallengeShindanmaker) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := ChimpoInsertionChallengeRegex.FindStringIndex(message.Content)
	result, err := cs.Client.Do(ctx, cs.Client.Name(message.Account), "https://shindanmaker.com/a/670773")
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
