package action

import (
	"regexp"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	DBRegex = regexp.MustCompile(`^SQL:\s?(.+)`)
)

type db struct {
	Environment config.Environment
}

func NewDB(environment config.Environment) service.Action {
	return &db{
		Environment: environment,
	}
}

func (d *db) Name() string {
	return "DB"
}

func (d *db) Target(message service.Message) bool {
	return !message.IsReblog &&
		message.Account.ID == d.Environment.Mastodon.UserID &&
		DBRegex.MatchString(message.Content)
}

func (d *db) Event(message service.Message) (service.Event, int, error) {
	index := DBRegex.FindStringIndex(message.Content)
	matches := DBRegex.FindStringSubmatch(message.Content)

	event := service.AdministrationEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Type:        d.Name(),
		Command:     matches[1],
	}

	return &event, index[0], nil
}
