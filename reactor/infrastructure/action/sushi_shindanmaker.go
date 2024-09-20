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
	SushiRegex  = regexp.MustCompile(`(🍣|寿司|すし|ちん(ちん|ぽ|こ))(握|にぎ)`)
	SushiEmojis = map[string]bool{
		"thinking_sushi":      true,
		"ios_big_sushi_1":     true,
		"ios_big_sushi_2":     true,
		"ios_big_sushi_3":     true,
		"ios_big_sushi_4":     true,
		"top_left_sushi":      true,
		"top_center_sushi":    true,
		"top_right_sushi":     true,
		"middle_left_sushi":   true,
		"middle_right_sushi":  true,
		"bottom_left_sushi":   true,
		"bottom_center_sushi": true,
		"bottom_right_sushi":  true,
	}
)

type sushiShindanmaker struct {
	Client         client.Shindanmaker
	MastodonUserID string
}

func NewSushiShindanmaker(c client.Shindanmaker, mastodonUserID string) service.Action {
	return &sushiShindanmaker{
		Client:         c,
		MastodonUserID: mastodonUserID,
	}
}

func (ss *sushiShindanmaker) Name() string {
	return "寿司職人"
}

func (ss *sushiShindanmaker) Target(message service.Message) bool {
	if message.IsReblog || (message.Account.ID == ss.MastodonUserID && message.InReplyToID != "") {
		return false
	}

	for _, emoji := range message.Emojis {
		if SushiEmojis[emoji.Shortcode] {
			return true
		}
	}

	return SushiRegex.MatchString(message.Content)
}

func (ss *sushiShindanmaker) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := SushiRegex.FindStringIndex(message.Content)
	if index == nil {
		index = make([]int, 2)
	}

	result, err := ss.Client.Do(ctx, ss.Client.Name(message.Account), "https://shindanmaker.com/a/577901")
	if err != nil {
		return nil, index[0], fmt.Errorf("failed to create event: %w", err)
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        io.NopCloser(strings.NewReader(result)),
		Visibility:  message.Visibility,
	}

	return event, index[0], nil
}
