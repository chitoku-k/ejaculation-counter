package action

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	BlueArchiveEcchiGameRegex = regexp.MustCompile(`((ブ|ﾌﾞ)[ルﾙ][アｱ][カｶ]|(ブ|ﾌﾞ)[ルﾙ][ーｰ][アｱ][ーｰ][カｶ][イｲ](ブ|ﾌﾞ))は?(えっ?ち|[エｴ][ッｯ]?[チﾁ])`)
)

type blueArchiveEcchiGameShindanmaker struct {
	Client      client.Shindanmaker
	Environment config.Environment
}

func NewBlueArchiveEcchiGameShindanmaker(c client.Shindanmaker, environment config.Environment) service.Action {
	return &blueArchiveEcchiGameShindanmaker{
		Client:      c,
		Environment: environment,
	}
}

func (bs *blueArchiveEcchiGameShindanmaker) Name() string {
	return "ブルアカはエッチなゲームではありません"
}

func (bs *blueArchiveEcchiGameShindanmaker) Target(message service.Message) bool {
	if message.IsReblog || (message.Account.ID == bs.Environment.Mastodon.UserID && message.InReplyToID != "") {
		return false
	}

	for _, tag := range message.Tags {
		if tag.Name == "shindanmaker" {
			return false
		}
	}

	return BlueArchiveEcchiGameRegex.MatchString(message.Content)
}

func (bs *blueArchiveEcchiGameShindanmaker) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := BlueArchiveEcchiGameRegex.FindStringSubmatchIndex(message.Content)
	result, err := bs.Client.Do(ctx, bs.Client.Name(message.Account), "https://shindanmaker.com/a/1111071")
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
