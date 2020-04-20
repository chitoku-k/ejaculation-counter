package client

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type through struct {
	Client http.Client
}

type Through interface {
	Do(targetURL string) (ThroughChallengeResult, error)
}

type ThroughChallengeResult []string

func NewThrough(client http.Client) Through {
	return &through{
		Client: client,
	}
}

func (t *through) Do(targetURL string) (ThroughChallengeResult, error) {
	res, err := t.Client.Get(targetURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch challenge result")
	}
	defer res.Body.Close()

	var result ThroughChallengeResult
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode challenge result")
	}

	return result, nil
}
