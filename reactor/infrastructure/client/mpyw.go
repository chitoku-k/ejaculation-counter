//go:generate go tool mockgen -source=mpyw.go -destination=mpyw_mock.go -package=client -self_package=github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client

package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type mpyw struct {
	Client *http.Client
}

type Mpyw interface {
	Do(ctx context.Context, targetURL string, count int) (io.ReadCloser, error)
}

func NewMpyw(client *http.Client) Mpyw {
	return &mpyw{
		Client: client,
	}
}

func (m *mpyw) Do(ctx context.Context, targetURL string, count int) (io.ReadCloser, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse given targetURL: %w", err)
	}

	q := u.Query()
	q.Add("count", fmt.Sprint(count))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create challenge request: %w", err)
	}

	res, err := m.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch challenge result: %w", err)
	}

	return res.Body, nil
}
