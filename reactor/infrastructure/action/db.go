package action

import (
	"context"
	"regexp"

	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	DBRegex = regexp.MustCompile(`^SQL:\s?(.+)`)
)

type db struct {
	MastodonUserID string
}

func NewDB(mastodonUserID string) service.Action {
	return &db{
		MastodonUserID: mastodonUserID,
	}
}

func (d *db) Name() string {
	return "DB"
}

func (d *db) Target(message service.Message) bool {
	return !message.IsReblog &&
		message.Account.ID == d.MastodonUserID &&
		DBRegex.MatchString(message.Content)
}

func (d *db) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := DBRegex.FindStringIndex(message.Content)
	matches := DBRegex.FindStringSubmatch(message.Content)

	if index == nil || matches == nil {
		return nil, 0, service.ErrNoMatch
	}

	event := service.AdministrationEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Type:        d.Name(),
		Command:     matches[1],
		Visibility:  message.Visibility,
	}

	return event, index[0], nil
}
