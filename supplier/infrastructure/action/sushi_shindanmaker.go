package action

import (
	"regexp"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/pkg/errors"
)

var (
	SushiRegex  = regexp.MustCompile(`(🍣|寿司|すし|ちん(ちん|ぽ|こ))(握|にぎ)`)
	SushiEmojis = map[string]bool{
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
	Client client.Shindanmaker
}

func NewSushiShindanmaker(c client.Shindanmaker) service.Action {
	return &sushiShindanmaker{
		Client: c,
	}
}

func (ss *sushiShindanmaker) Name() string {
	return "寿司職人"
}

func (ss *sushiShindanmaker) Target(message service.Message) bool {
	if message.IsReblog {
		return false
	}

	for _, emoji := range message.Emojis {
		if SushiEmojis[emoji.Shortcode] {
			return true
		}
	}

	return SushiRegex.MatchString(message.Content)
}

func (ss *sushiShindanmaker) Event(message service.Message) (service.Event, int, error) {
	index := SushiRegex.FindStringIndex(message.Content)
	if index == nil {
		index = make([]int, 2)
	}

	result, err := ss.Client.Do(ss.Client.Name(message.Account), "https://shindanmaker.com/a/577901")
	if err != nil {
		return nil, index[0], errors.Wrap(err, "failed to create event")
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        result,
		Visibility:  message.Visibility,
	}

	return &event, index[0], nil
}
