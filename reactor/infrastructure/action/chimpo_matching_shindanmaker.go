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
	ChimpoMatchingRegex = regexp.MustCompile(`ちん(ちん|ぽ|こ)(揃|そろ)え`)
)

type chimpoMatchingShindanmaker struct {
	Client         client.Shindanmaker
	MastodonUserID string
}

func NewChimpoMatchingShindanmaker(c client.Shindanmaker, mastodonUserID string) service.Action {
	return &chimpoMatchingShindanmaker{
		Client:         c,
		MastodonUserID: mastodonUserID,
	}
}

func (bs *chimpoMatchingShindanmaker) Name() string {
	return "ちんぽ揃えゲーム"
}

func (bs *chimpoMatchingShindanmaker) Target(message service.Message) bool {
	if message.IsReblog || (message.Account.ID == bs.MastodonUserID && message.InReplyToID != "") {
		return false
	}

	for _, tag := range message.Tags {
		if tag.Name == "ちんぽ揃えゲーム" {
			return false
		}
	}

	return ChimpoMatchingRegex.MatchString(message.Content)
}

func (bs *chimpoMatchingShindanmaker) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := ChimpoMatchingRegex.FindStringIndex(message.Content)
	result, err := bs.Client.Do(ctx, bs.Client.Name(message.Account), "https://shindanmaker.com/a/855159")
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
