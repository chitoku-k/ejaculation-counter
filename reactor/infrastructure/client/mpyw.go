//go:generate mockgen -source=mpyw.go -destination=mpyw_mock.go -package=client -self_package=github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client

package client

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/wrapper"
)

type mpyw struct {
	Client wrapper.HttpClient
}

type Mpyw interface {
	Do(targetURL string, count int) (MpywChallengeResult, error)
}

type MpywChallengeResult struct {
	Title  string   `json:"title"`
	Result []string `json:"result"`
}

func NewMpyw(client wrapper.HttpClient) Mpyw {
	return &mpyw{
		Client: client,
	}
}

func (m *mpyw) Do(targetURL string, count int) (MpywChallengeResult, error) {
	var result MpywChallengeResult

	u, err := url.Parse(targetURL)
	if err != nil {
		return result, fmt.Errorf("failed to parse given targetURL: %w", err)
	}

	q := u.Query()
	q.Add("count", fmt.Sprint(count))
	u.RawQuery = q.Encode()

	res, err := m.Client.Get(u.String())
	if err != nil {
		return result, fmt.Errorf("failed to fetch challenge result: %w", err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("failed to decode challenge result: %w", err)
	}

	return result, nil
}
