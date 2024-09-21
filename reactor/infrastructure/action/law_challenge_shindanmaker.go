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
	LawChallengeRegex = regexp.MustCompile(`法律((ギ|ｷﾞ)[リﾘ](ギ|ｷﾞ)[リﾘ])?[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)`)
)

type lawChallengeShindanmaker struct {
	Client         client.Shindanmaker
	MastodonUserID string
}

func NewLawChallengeShindanmaker(c client.Shindanmaker, mastodonUserID string) service.Action {
	return &lawChallengeShindanmaker{
		Client:         c,
		MastodonUserID: mastodonUserID,
	}
}

func (ls *lawChallengeShindanmaker) Name() string {
	return "法律ギリギリチャレンジ"
}

func (ls *lawChallengeShindanmaker) Target(message service.Message) bool {
	if message.IsReblog || (message.Account.ID == ls.MastodonUserID && message.InReplyToID != "") {
		return false
	}

	for _, tag := range message.Tags {
		if tag.Name == "法律ギリギリチャレンジ" {
			return false
		}
	}

	return LawChallengeRegex.MatchString(message.Content)
}

func (ls *lawChallengeShindanmaker) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := LawChallengeRegex.FindStringIndex(message.Content)
	result, err := ls.Client.Do(ctx, ls.Client.Name(message.Account), "https://shindanmaker.com/a/877845")
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
