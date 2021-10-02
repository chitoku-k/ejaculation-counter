//go:generate mockgen -source=doublet.go -destination=doublet_mock.go -package=client -self_package=github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type doublet struct {
	Client *http.Client
}

type Doublet interface {
	Do(ctx context.Context, targetURL string) (DoubletResult, error)
}

type DoubletResult []string

func NewDoublet(client *http.Client) Doublet {
	return &doublet{
		Client: client,
	}
}

func (t *doublet) Do(ctx context.Context, targetURL string) (DoubletResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create challenge request: %w", err)
	}

	res, err := t.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch challenge result: %w", err)
	}
	defer res.Body.Close()

	var result DoubletResult
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode challenge result: %w", err)
	}

	return result, nil
}
