package action

import (
	"context"
	"io"
	"math/rand"
	"regexp"
	"strconv"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/reader"
	"github.com/chitoku-k/ejaculation-counter/reactor/repository"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	DoubletRegex = regexp.MustCompile(`今日の\s*(?:doublet|二重語)|(?:\s*(\d+)\s*連\s*)?(?:doublet|二重語)\s*(?:が|ガ|ｶﾞ)[チﾁ][ャｬ]`)
)

type doublet struct {
	Repository  repository.DoubletRepository
	Environment config.Environment
}

func NewDoublet(repository repository.DoubletRepository, environment config.Environment) service.Action {
	return &doublet{
		Repository:  repository,
		Environment: environment,
	}
}

func (d *doublet) Name() string {
	return "二重語ガチャ"
}

func (d *doublet) Target(message service.Message) bool {
	return !message.IsReblog &&
		(message.Account.ID != d.Environment.Mastodon.UserID || message.InReplyToID == "") &&
		DoubletRegex.MatchString(message.Content)
}

func (d *doublet) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := DoubletRegex.FindStringIndex(message.Content)
	matches := DoubletRegex.FindStringSubmatch(message.Content)

	count, err := strconv.Atoi(matches[1])
	if err != nil {
		count = 1
	}

	items := d.Repository.Get()
	generator := func() string {
		return items[rand.Intn(len(items))]
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        io.NopCloser(reader.NewStringFuncReader("\n", count, generator)),
		Visibility:  message.Visibility,
	}

	return event, index[0], nil
}
