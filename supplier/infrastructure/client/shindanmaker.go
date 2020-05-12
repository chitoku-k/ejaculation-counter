//go:generate mockgen -source=shindanmaker.go -destination=shindanmaker_mock.go -package=client -self_package=github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client

package client

import (
	"bytes"
	"html"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/pkg/errors"
)

var (
	NameRegex          = regexp.MustCompile(`[$\\]{?\d+`)
	ShindanNameRegex   = regexp.MustCompile(`[@＠].+|[\(（].+[\)）]`)
	ShindanResultRegex = regexp.MustCompile(`<textarea id="copy_text_140"(?:[^>]+)>([\s\S]*?)<\/textarea>`)
)

type shindanmaker struct {
	Client wrapper.HttpClient
}

type Shindanmaker interface {
	Do(name string, targerURL string) (string, error)
	Name(account service.Account) string
}

func NewShindanmaker(client wrapper.HttpClient) Shindanmaker {
	return &shindanmaker{
		Client: client,
	}
}

func (s *shindanmaker) Name(account service.Account) string {
	if account.DisplayName != "" {
		return account.DisplayName
	} else {
		return account.Username
	}
}

func (s *shindanmaker) Do(name string, targetURL string) (string, error) {
	name = NameRegex.ReplaceAllString(name, "\\$0")

	if name != strings.Join(ShindanNameRegex.FindStringSubmatch(name), "") {
		name = ShindanNameRegex.ReplaceAllString(name, "")
	}

	values := url.Values{}
	values.Add("u", name)

	res, err := s.Client.Post(targetURL, "application/x-www-form-urlencoded", strings.NewReader(values.Encode()))
	if err != nil {
		return "", errors.Wrap(err, "failed to fetch shindan result")
	}
	defer res.Body.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, res.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read shindan result")
	}

	matches := ShindanResultRegex.FindSubmatch(buf.Bytes())
	if matches == nil {
		return "", errors.New("failed to parse shindan result")
	}

	return html.UnescapeString(string(matches[1])), nil
}
