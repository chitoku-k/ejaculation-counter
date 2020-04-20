package action

import (
	"math/rand"
	"regexp"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/supplier/service"
)

var (
	Ofuton      = []string{"お", "ふ", "と", "ん"}
	OfutonRegex = regexp.MustCompile(`ふとん[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)`)
)

type ofutonChallenge struct{}

func NewOfufutonChallenge() service.Action {
	return &ofutonChallenge{}
}

func (*ofutonChallenge) Name() string {
	return "おふとんチャレンジ"
}

func (*ofutonChallenge) Target(message service.Message) bool {
	if message.IsReblog {
		return false
	}

	for _, tag := range message.Tags {
		if tag.Name == "おふとんチャレンジ" {
			return false
		}
	}

	return OfutonRegex.MatchString(message.Content)
}

func (*ofutonChallenge) Event(message service.Message) (service.Event, int, error) {
	index := OfutonRegex.FindStringIndex(message.Content)
	ofuton := make([]string, len(Ofuton))

	for i := 0; i < len(Ofuton); i++ {
		ofuton = append(ofuton, Ofuton[rand.Intn(len(Ofuton))])
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        strings.Join(ofuton, "") + "\n#おふとんチャレンジ",
		Visibility:  message.Visibility,
	}

	return &event, index[0], nil
}
