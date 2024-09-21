package action

import (
	"context"
	"regexp"

	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	PyuUpdateRegex = regexp.MustCompile(`^ぴゅっ♡+$`)
)

type pyuUpdate struct {
	MastodonUserID string
}

func NewPyuUpdate(mastodonUserID string) service.Action {
	return &pyuUpdate{
		MastodonUserID: mastodonUserID,
	}
}

func (pu *pyuUpdate) Name() string {
	return "ぴゅっ♡"
}

func (pu *pyuUpdate) Target(message service.Message) bool {
	return !message.IsReblog &&
		message.Account.ID == pu.MastodonUserID &&
		PyuUpdateRegex.MatchString(message.Content)
}

func (pu *pyuUpdate) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := PyuUpdateRegex.FindStringIndex(message.Content)
	event := service.IncrementEvent{
		Year:  message.CreatedAt.Year(),
		Month: int(message.CreatedAt.Month()),
		Day:   message.CreatedAt.Day(),
	}

	return event, index[0], nil
}
