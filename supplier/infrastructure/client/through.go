//go:generate mockgen -source=through.go -destination=through_mock.go -package=client -self_package=github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client

package client

import (
	"encoding/json"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper"
	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "failed to fetch challenge result")
	}
	defer res.Body.Close()

	var result ThroughResult
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode challenge result")
	}

	return result, nil
}
