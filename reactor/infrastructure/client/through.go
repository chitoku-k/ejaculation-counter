//go:generate mockgen -source=through.go -destination=through_mock.go -package=client -self_package=github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client

package client

import (
	"encoding/json"
	"fmt"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/wrapper"
)

type through struct {
	Client wrapper.HttpClient
}

type Through interface {
	Do(targetURL string) (ThroughResult, error)
}

type ThroughResult []string

func NewThrough(client wrapper.HttpClient) Through {
	return &through{
		Client: client,
	}
}

func (t *through) Do(targetURL string) (ThroughResult, error) {
	res, err := t.Client.Get(targetURL)
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
