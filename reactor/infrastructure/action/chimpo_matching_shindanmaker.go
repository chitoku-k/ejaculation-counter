package action

import (
	"fmt"
	"regexp"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	ChimpoMatchingRegex = regexp.MustCompile(`ちん(ちん|ぽ|こ)(揃|そろ)え`)
)

type chimpoMatchingShindanmaker struct {
	Client      client.Shindanmaker
	Environment config.Environment
}

func NewChimpoMatchingShindanmaker(c client.Shindanmaker, environment config.Environment) service.Action {
	return &chimpoMatchingShindanmaker{
		Client:      c,
		Environment: environment,
	}
}

func (bs *chimpoMatchingShindanmaker) Name() string {
	return "ちんぽ揃えゲーム"
}

func (bs *chimpoMatchingShindanmaker) Target(message service.Message) bool {
	if message.IsReblog || (message.Account.ID == bs.Environment.Mastodon.UserID && message.InReplyToID != "") {
		return false
	}

	for _, tag := range message.Tags {
		if tag.Name == "ちんぽ揃えゲーム" {
			return false
		}
	}

	return ChimpoMatchingRegex.MatchString(message.Content)
}

func (bs *chimpoMatchingShindanmaker) Event(message service.Message) (service.Event, int, error) {
	index := ChimpoMatchingRegex.FindStringIndex(message.Content)
	result, err := bs.Client.Do(bs.Client.Name(message.Account), "https://shindanmaker.com/a/855159")
	if err != nil {
		return nil, index[0], fmt.Errorf("failed to create event: %w", err)
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        result,
		Visibility:  message.Visibility,
	}

	return &event, index[0], nil
}
