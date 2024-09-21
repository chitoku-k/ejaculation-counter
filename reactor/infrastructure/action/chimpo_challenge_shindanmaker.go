package action

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	ChimpoChallengeRegex = regexp.MustCompile(`ちん(ちん|ぽ|こ)[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)`)
)

type chimpoChallengeShindanmaker struct {
	Client         client.Shindanmaker
	MastodonUserID string
}

func NewChimpoChallengeShindanmaker(c client.Shindanmaker, mastodonUserID string) service.Action {
	return &chimpoChallengeShindanmaker{
		Client:         c,
		MastodonUserID: mastodonUserID,
	}
}

func (cs *chimpoChallengeShindanmaker) Name() string {
	return "ちんぽチャレンジ"
}

func (cs *chimpoChallengeShindanmaker) Target(message service.Message) bool {
	if message.IsReblog || (message.Account.ID == cs.MastodonUserID && message.InReplyToID != "") {
		return false
	}

	for _, tag := range message.Tags {
		if tag.Name == "ちんぽチャレンジ" {
			return false
		}
	}

	return ChimpoChallengeRegex.MatchString(message.Content)
}

func (cs *chimpoChallengeShindanmaker) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := ChimpoChallengeRegex.FindStringIndex(message.Content)
	result, err := cs.Client.Do(ctx, cs.Client.Name(message.Account), "https://shindanmaker.com/a/656461")
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
