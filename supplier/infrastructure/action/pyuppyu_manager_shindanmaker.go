package action

import (
	"fmt"
	"regexp"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
)

var (
	PyuppyuManagerRegex = regexp.MustCompile(`ぴゅっぴゅしても?[いよ良]い[?？]`)
)

type pyuppyuManagerShindanmaker struct {
	Client client.Shindanmaker
}

func NewPyuppyuManagerShindanmaker(c client.Shindanmaker) service.Action {
	return &pyuppyuManagerShindanmaker{
		Client: c,
	}
}

func (ps *pyuppyuManagerShindanmaker) Name() string {
	return "おちんちんぴゅっぴゅ管理官の毎日"
}

func (ps *pyuppyuManagerShindanmaker) Target(message service.Message) bool {
	return !message.IsReblog && PyuppyuManagerRegex.MatchString(message.Content)
}

func (ps *pyuppyuManagerShindanmaker) Event(message service.Message) (service.Event, int, error) {
	index := PyuppyuManagerRegex.FindStringIndex(message.Content)
	result, err := ps.Client.Do(ps.Client.Name(message.Account), "https://shindanmaker.com/a/503598")
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
