//go:generate mockgen -source=shindanmaker.go -destination=shindanmaker_mock.go -package=client -self_package=github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client

package client

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/chitoku-k/ejaculation-counter/reactor/service"
)

const (
	UserAgent = "Mozilla/5.0 (compatible)"
)

var (
	NameRegex          = regexp.MustCompile(`[$\\]{?\d+`)
	ShindanNameRegex   = regexp.MustCompile(`[@＠].+|[\(（].+[\)）]`)
	ShindanTokenRegex  = regexp.MustCompile(`<meta[^>]*name="csrf-token"[^>]*content="([\s\S]*?)"[^>]*>`)
	ShindanResultRegex = regexp.MustCompile(`<textarea[^>]*id="copy-textarea-140"[^>]*>([\s\S]*?)<\/textarea>`)
)

type shindanmaker struct {
	Client *http.Client
}

type Shindanmaker interface {
	Do(ctx context.Context, name string, targerURL string) (string, error)
	Name(account service.Account) string
}

func NewShindanmaker(client *http.Client) Shindanmaker {
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

func (s *shindanmaker) token(ctx context.Context, targetURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create sindan page request: %w", err)
	}

	req.Header.Set("Accept", "*")
	req.Header.Set("User-Agent", UserAgent)

	res, err := s.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch shindan page: %w", err)
	}
	defer res.Body.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read shindan page: %w", err)
	}

	matches := ShindanTokenRegex.FindSubmatch(buf.Bytes())
	if matches == nil {
		return "", fmt.Errorf("failed to parse shindan page")
	}

	return string(matches[1]), nil
}

func (s *shindanmaker) Do(ctx context.Context, name string, targetURL string) (string, error) {
	token, err := s.token(ctx, targetURL)
	if err != nil {
		return "", err
	}

	name = NameRegex.ReplaceAllString(name, "\\$0")
	if name != strings.Join(ShindanNameRegex.FindStringSubmatch(name), "") {
		name = ShindanNameRegex.ReplaceAllString(name, "")
	}

	values := url.Values{
		"type":        []string{"name"},
		"shindanName": []string{name},
		"_token":      []string{token},
	}
	body := strings.NewReader(values.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.ReplaceAll(targetURL, "/a/", "/"), body)
	if err != nil {
		return "", fmt.Errorf("failed to create shindan result request: %w", err)
	}
	req.Header.Set("Accept", "*")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", UserAgent)

	res, err := s.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch shindan result: %w", err)
	}
	defer res.Body.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read shindan result: %w", err)
	}

	matches := ShindanResultRegex.FindSubmatch(buf.Bytes())
	if matches == nil {
		return "", fmt.Errorf("failed to parse shindan result")
	}

	return strings.ReplaceAll(
		html.UnescapeString(string(matches[1])),
		"\u2002",
		" ",
	), nil
}
