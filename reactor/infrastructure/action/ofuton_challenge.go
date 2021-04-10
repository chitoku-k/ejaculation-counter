package action

import (
	"regexp"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/wrapper"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	Ofuton      = []string{"お", "ふ", "と", "ん"}
	OfutonRegex = regexp.MustCompile(`ふとん[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)`)
)

type ofutonChallenge struct {
	Random wrapper.Random
}

func NewOfufutonChallenge(r wrapper.Random) service.Action {
	return &ofutonChallenge{r}
}

func (oc *ofutonChallenge) Name() string {
	return "おふとんチャレンジ"
}

func (oc *ofutonChallenge) Target(message service.Message) bool {
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

func (oc *ofutonChallenge) Event(message service.Message) (service.Event, int, error) {
	index := OfutonRegex.FindStringIndex(message.Content)
	ofuton := make([]string, len(Ofuton))

	for i := 0; i < len(Ofuton); i++ {
		ofuton = append(ofuton, Ofuton[oc.Random.Intn(len(Ofuton))])
	}

	event := service.ReplyEvent{
		InReplyToID: message.ID,
		Acct:        message.Account.Acct,
		Body:        strings.Join(ofuton, "") + "\n#おふとんチャレンジ",
		Visibility:  message.Visibility,
	}

	return &event, index[0], nil
}
