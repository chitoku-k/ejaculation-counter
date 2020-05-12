package action

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/pkg/errors"
)

var (
	MpywRegex = regexp.MustCompile(`(?:mpyw\s*|まっぴー|実務経験)(?:\s*(\d+)\s*連)?(?:が|ガ|ｶﾞ)[チﾁ][ャｬ]`)
)

type mpyw struct {
	Client client.Mpyw
}

func NewMpyw(c client.Mpyw) service.Action {
	return &mpyw{
		Client: c,
	}
}

func (m *mpyw) Name() string {
	return "実務経験ガチャ"
}

func (m *mpyw) Target(message service.Message) bool {
	return !message.IsReblog && MpywRegex.MatchString(message.Content)
}

func (m *mpyw) Event(message service.Message) (service.Event, int, error) {
	index := MpywRegex.FindStringIndex(message.Content)
	matches := MpywRegex.FindStringSubmatch(message.Content)

	count, err := strconv.Atoi(matches[1])
	if err != nil {
		count = 1
	}

	result, err := m.Client.Do("https://mpyw.kb10uy.org/api", count)
	if err != nil {
		return nil, index[0], errors.Wrap(err, "failed to create event")
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        strings.Join(result.Result, "\n"),
		Visibility:  message.Visibility,
	}

	return &event, index[0], nil
}
