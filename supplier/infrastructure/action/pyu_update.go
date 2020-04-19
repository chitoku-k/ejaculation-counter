package action

import (
	"regexp"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
)

var (
	PyuUpdateRegex = regexp.MustCompile(`^ぴゅっ♡+$`)
)

type pyuUpdate struct {
	Environment config.Environment
}

func NewPyuUpdateShindanmaker(environment config.Environment) service.Action {
	return &pyuUpdate{
		Environment: environment,
	}
}

func (pu *pyuUpdate) Name() string {
	return "ぴゅっ♡"
}

func (pu *pyuUpdate) Target(message service.Message) bool {
	return !message.IsReblog &&
		message.Account.ID == pu.Environment.Mastodon.UserID &&
		PyuUpdateRegex.MatchString(message.Content)
}

func (pu *pyuUpdate) Event(message service.Message) (service.Event, int, error) {
	index := PyuUpdateRegex.FindStringIndex(message.Content)
	event := service.IncrementEvent{
		UserID: message.Account.ID,
		Date:   message.CreatedAt.String(),
	}

	return &event, index[0], nil
}
