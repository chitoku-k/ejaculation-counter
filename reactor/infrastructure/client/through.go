//go:generate mockgen -source=through.go -destination=through_mock.go -package=client -self_package=github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type through struct {
	Client *http.Client
}

type Through interface {
	Do(ctx context.Context, targetURL string) (ThroughResult, error)
}

type ThroughResult []string

func NewThrough(client *http.Client) Through {
	return &through{
		Client: client,
	}
}

func (t *through) Do(ctx context.Context, targetURL string) (ThroughResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create challenge request: %w", err)
	}

	res, err := t.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch challenge result: %w", err)
	}
	defer res.Body.Close()

	var result ThroughResult
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode challenge result: %w", err)
	}

	return result, nil
}
