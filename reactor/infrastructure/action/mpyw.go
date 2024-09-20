package action

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/reader"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	MpywRegex = regexp.MustCompile(`(?:mpyw\s*|まっぴー|実務経験)(?:\s*(\d+)\s*連)?(?:が|ガ|ｶﾞ)[チﾁ][ャｬ]`)
)

type mpyw struct {
	Client         client.Mpyw
	MastodonUserID string
	MpywAPIURL     string
}

func NewMpyw(c client.Mpyw, mastodonUserID, mpywAPIURL string) service.Action {
	return &mpyw{
		Client:         c,
		MastodonUserID: mastodonUserID,
		MpywAPIURL:     mpywAPIURL,
	}
}

func (m *mpyw) Name() string {
	return "実務経験ガチャ"
}

func (m *mpyw) Target(message service.Message) bool {
	return !message.IsReblog &&
		(message.Account.ID != m.MastodonUserID || message.InReplyToID == "") &&
		MpywRegex.MatchString(message.Content)
}

func (m *mpyw) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := MpywRegex.FindStringIndex(message.Content)
	matches := MpywRegex.FindStringSubmatch(message.Content)

	count, err := strconv.Atoi(matches[1])
	if err != nil {
		count = 1
	}

	result, err := m.Client.Do(ctx, m.MpywAPIURL, count)
	if err != nil {
		return nil, index[0], fmt.Errorf("failed to create event: %w", err)
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        reader.NewJsonStreamReader("\n", 2, result),
		Visibility:  message.Visibility,
	}

	return event, index[0], nil
}
