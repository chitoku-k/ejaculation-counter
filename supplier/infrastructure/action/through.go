package action

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
)

var (
	ThroughRegex = regexp.MustCompile(`(?:\s*(\d+)\s*連)?駿河茶|今日の\s*through|through\s*(?:が|ガ|ｶﾞ)[チﾁ][ャｬ]`)
)

type through struct {
	Client      client.Through
	Environment config.Environment
}

func NewThrough(c client.Through, environment config.Environment) service.Action {
	return &through{
		Client:      c,
		Environment: environment,
	}
}

func (t *through) Name() string {
	return "駿河茶"
}

func (t *through) Target(message service.Message) bool {
	return !message.IsReblog && ThroughRegex.MatchString(message.Content)
}

func (t *through) Event(message service.Message) (service.Event, int, error) {
	index := ThroughRegex.FindStringIndex(message.Content)
	matches := ThroughRegex.FindStringSubmatch(message.Content)

	count, err := strconv.Atoi(matches[1])
	if err != nil {
		count = 1
	}

	items, err := t.Client.Do(fmt.Sprintf("http://%s/through", t.Environment.Reactor.Host))
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
