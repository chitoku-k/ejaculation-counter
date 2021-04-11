package action

import (
	"fmt"
	"regexp"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	BattleChimpoRegex = regexp.MustCompile(`ちん(ちん|ぽ|こ)(なん[かぞ])?に([勝か][たちつてと]|[負ま][かきくけこ])`)
)

type battleChimpoShindanmaker struct {
	Client      client.Shindanmaker
	Environment config.Environment
}

func NewBattleChimpoShindanmaker(c client.Shindanmaker, environment config.Environment) service.Action {
	return &battleChimpoShindanmaker{
		Client:      c,
		Environment: environment,
	}
}

func (bs *battleChimpoShindanmaker) Name() string {
	return "絶対おちんぽなんかに負けない！"
}

func (bs *battleChimpoShindanmaker) Target(message service.Message) bool {
	return !message.IsReblog &&
		(message.Account.ID != bs.Environment.Mastodon.UserID || message.InReplyToID == "") &&
		BattleChimpoRegex.MatchString(message.Content)
}

func (bs *battleChimpoShindanmaker) Event(message service.Message) (service.Event, int, error) {
	index := BattleChimpoRegex.FindStringIndex(message.Content)
	result, err := bs.Client.Do(bs.Client.Name(message.Account), "https://shindanmaker.com/a/584238")
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
