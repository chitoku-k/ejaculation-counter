package action

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	OfutonManagerRegex = regexp.MustCompile(`ふとん(し|(入|はい|い|行|潜|もぐ)っ)ても?[いよ良]い[?？]`)
	OfutonRules        = []OfutonRule{
		{"しこしこして", "もふもふさせて"},
		{"しこしこ", "もふもふ"},
		{"しゅっしゅ", "もふもふ"},
		{"ぴゅっぴゅって", "もふもふって"},
		{"ぴゅっぴゅ", "おふとん"},
		{"いじるの", "おふとん"},
		{"おちんちん", "おふとん"},
		{"ちんちん", "おふとん"},
		{"出せる", "もふもふできる"},
		{"出し", "もふもふし"},
		{"手の平に", "朝まで"},
	}
)

type OfutonRule struct {
	Before string
	After  string
}

type ofutonManagerShindanmaker struct {
	Client         client.Shindanmaker
	MastodonUserID string
}

func NewOfutonManagerShindanmaker(c client.Shindanmaker, mastodonUserID string) service.Action {
	return &ofutonManagerShindanmaker{
		Client:         c,
		MastodonUserID: mastodonUserID,
	}
}

func (os *ofutonManagerShindanmaker) Name() string {
	return "おふとん管理官の毎日"
}

func (os *ofutonManagerShindanmaker) Target(message service.Message) bool {
	return !message.IsReblog &&
		(message.Account.ID != os.MastodonUserID || message.InReplyToID == "") &&
		OfutonManagerRegex.MatchString(message.Content)
}

func (os *ofutonManagerShindanmaker) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := OfutonManagerRegex.FindStringIndex(message.Content)
	result, err := os.Client.Do(ctx, os.Client.Name(message.Account), "https://shindanmaker.com/a/503598")
	if err != nil {
		return nil, index[0], fmt.Errorf("failed to create event: %w", err)
	}

	for _, rule := range OfutonRules {
		result = strings.ReplaceAll(result, rule.Before, rule.After)
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        io.NopCloser(strings.NewReader(result)),
		Visibility:  message.Visibility,
	}

	return event, index[0], nil
}
