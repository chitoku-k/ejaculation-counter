package client

import (
	"bytes"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/pkg/errors"
)

var (
	NameRegex          = regexp.MustCompile(`([$\\]{?\d+)`)
	ShindanNameRegex   = regexp.MustCompile(`(@.+|[\(（].+[\)）])$`)
	ShindanResultRegex = regexp.MustCompile(`<textarea id="copy_text_140"(?:[^>]+)>([\s\S]*?)<\/textarea>`)
)

type shindanmaker struct {
	Client http.Client
}

type Shindanmaker interface {
	Do(name string, targerURL string) (string, error)
	Name(account service.Account) string
}

func NewShindanmaker(client http.Client) Shindanmaker {
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
	name = NameRegex.ReplaceAllString(name, "\\$1")
	name = ShindanNameRegex.ReplaceAllString(name, "")

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

	return html.UnescapeString(
		html.UnescapeString(
			string(matches[1]),
		),
	), nil
}
