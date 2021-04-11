//go:generate mockgen -source=shindanmaker.go -destination=shindanmaker_mock.go -package=client -self_package=github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client

package client

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/wrapper"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

var (
	NameRegex          = regexp.MustCompile(`[$\\]{?\d+`)
	ShindanNameRegex   = regexp.MustCompile(`[@＠].+|[\(（].+[\)）]`)
	ShindanTokenRegex  = regexp.MustCompile(`<meta[^>]*name="csrf-token"[^>]*content="([\s\S]*?)"[^>]*>`)
	ShindanResultRegex = regexp.MustCompile(`<textarea[^>]*id="copy-textarea-140"[^>]*>([\s\S]*?)<\/textarea>`)
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
	top, err := s.Client.Get(targetURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch shindan top: %w", err)
	}
	defer top.Body.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, top.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read shindan top: %w", err)
	}

	matches := ShindanTokenRegex.FindSubmatch(buf.Bytes())
	if matches == nil {
		return "", fmt.Errorf("failed to parse shindan top")
	}

	name = NameRegex.ReplaceAllString(name, "\\$0")

	if name != strings.Join(ShindanNameRegex.FindStringSubmatch(name), "") {
		name = ShindanNameRegex.ReplaceAllString(name, "")
	}

	values := url.Values{}
	values.Add("name", name)
	values.Add("_token", string(matches[1]))

	res, err := s.Client.PostForm(strings.ReplaceAll(targetURL, "/a/", "/"), values)
	if err != nil {
		return "", fmt.Errorf("failed to fetch shindan result: %w", err)
	}
	defer res.Body.Close()

	buf.Reset()
	_, err = io.Copy(&buf, res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read shindan result: %w", err)
	}

	matches = ShindanResultRegex.FindSubmatch(buf.Bytes())
	if matches == nil {
		return "", fmt.Errorf("failed to parse shindan result")
	}

	return strings.ReplaceAll(
		html.UnescapeString(string(matches[1])),
		"\u2002",
		" ",
	), nil
}
