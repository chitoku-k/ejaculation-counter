package action

import (
	"context"
	"io"
	"regexp"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/reader"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	Ofuton      = []string{"お", "ふ", "と", "ん"}
	OfutonRegex = regexp.MustCompile(`ふとん[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)`)
)

type ofutonChallenge struct {
	Random         Random
	MastodonUserID string
}

func NewOfufutonChallenge(random Random, mastodonUserID string) service.Action {
	return &ofutonChallenge{
		Random:         random,
		MastodonUserID: mastodonUserID,
	}
}

func (oc *ofutonChallenge) Name() string {
	return "おふとんチャレンジ"
}

func (oc *ofutonChallenge) Target(message service.Message) bool {
	if message.IsReblog || (message.Account.ID == oc.MastodonUserID && message.InReplyToID != "") {
		return false
	}

	for _, tag := range message.Tags {
		if tag.Name == "おふとんチャレンジ" {
			return false
		}
	}

	return OfutonRegex.MatchString(message.Content)
}

func (oc *ofutonChallenge) Event(ctx context.Context, message service.Message) (service.Event, int, error) {
	index := OfutonRegex.FindStringIndex(message.Content)
	generator := func() string {
		return Ofuton[oc.Random.IntN(len(Ofuton))]
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body: io.NopCloser(
			io.MultiReader(
				reader.NewStringFuncReader("", len(Ofuton), generator),
				strings.NewReader("\n#おふとんチャレンジ"),
			),
		),
		Visibility: message.Visibility,
	}

	return event, index[0], nil
}
