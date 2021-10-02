package action

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	DoubletRegex = regexp.MustCompile(`今日の\s*(?:doublet|二重語)|(?:\s*(\d+)\s*連\s*)?(?:doublet|二重語)\s*(?:が|ガ|ｶﾞ)[チﾁ][ャｬ]`)
)

type doublet struct {
	Client      client.Doublet
	Environment config.Environment
}

func NewDoublet(c client.Doublet, environment config.Environment) service.Action {
	return &doublet{
		Client:      c,
		Environment: environment,
	}
}

func (t *doublet) Name() string {
	return "二重語ガチャ"
}

func (t *doublet) Target(message service.Message) bool {
	return !message.IsReblog &&
		(message.Account.ID != t.Environment.Mastodon.UserID || message.InReplyToID == "") &&
		DoubletRegex.MatchString(message.Content)
}

func (t *doublet) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := DoubletRegex.FindStringIndex(message.Content)
	matches := DoubletRegex.FindStringSubmatch(message.Content)

	count, err := strconv.Atoi(matches[1])
	if err != nil {
		count = 1
	}

	items, err := t.Client.Do(ctx, "http://localhost/doublet")
	if err != nil {
		return nil, index[0], fmt.Errorf("failed to create event: %w", err)
	}

	var result []string
	for i := 0; i < count; i++ {
		result = append(result, items[rand.Intn(len(items))])
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        strings.Join(result, "\n"),
		Visibility:  message.Visibility,
	}

	return &event, index[0], nil
}
