package action

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	MpywRegex = regexp.MustCompile(`(?:mpyw\s*|まっぴー|実務経験)(?:\s*(\d+)\s*連)?(?:が|ガ|ｶﾞ)[チﾁ][ャｬ]`)
)

type mpyw struct {
	Client      client.Mpyw
	Environment config.Environment
}

func NewMpyw(c client.Mpyw, environment config.Environment) service.Action {
	return &mpyw{
		Client:      c,
		Environment: environment,
	}
}

func (m *mpyw) Name() string {
	return "実務経験ガチャ"
}

func (m *mpyw) Target(message service.Message) bool {
	return !message.IsReblog &&
		(message.Account.ID != m.Environment.Mastodon.UserID || message.InReplyToID == "") &&
		MpywRegex.MatchString(message.Content)
}

func (m *mpyw) Event(message service.Message) (service.Event, int, error) {
	index := MpywRegex.FindStringIndex(message.Content)
	matches := MpywRegex.FindStringSubmatch(message.Content)

	count, err := strconv.Atoi(matches[1])
	if err != nil {
		count = 1
	}

	result, err := m.Client.Do(m.Environment.External.MpywAPIURL, count)
	if err != nil {
		return nil, index[0], fmt.Errorf("failed to create event: %w", err)
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        strings.Join(result.Result, "\n"),
		Visibility:  message.Visibility,
	}

	return &event, index[0], nil
}
