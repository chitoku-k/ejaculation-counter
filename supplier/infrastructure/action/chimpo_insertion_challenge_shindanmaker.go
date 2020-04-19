package action

import (
	"regexp"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/pkg/errors"
)

var (
	ChimpoInsertionChallengeRegex = regexp.MustCompile(`ちん(ちん|ぽ|こ)挿入[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)`)
)

type chimpoInsertionChallengeShindanmaker struct {
	Client client.Shindanmaker
}

func NewChimpoInsertionChallengeShindanmaker(c client.Shindanmaker) service.Action {
	return &chimpoInsertionChallengeShindanmaker{
		Client: c,
	}
}

func (cs *chimpoInsertionChallengeShindanmaker) Name() string {
	return "おちんぽ挿入チャレンジ"
}

func (cs *chimpoInsertionChallengeShindanmaker) Target(message service.Message) bool {
	if message.IsReblog {
		return false
	}

	for _, tag := range message.Tags {
		if tag.Name == "おちんぽ挿入チャレンジ" {
			return false
		}
	}

	return ChimpoInsertionChallengeRegex.MatchString(message.Content)
}

func (cs *chimpoInsertionChallengeShindanmaker) Event(message service.Message) (service.Event, int, error) {
	index := ChimpoInsertionChallengeRegex.FindStringIndex(message.Content)
	result, err := cs.Client.Do(cs.Client.Name(message.Account), "https://shindanmaker.com/a/670773")
	if err != nil {
		return nil, index[0], errors.Wrap(err, "failed to create event")
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        result,
		Visibility:  message.Visibility,
	}

	return &event, index[0], nil
}
