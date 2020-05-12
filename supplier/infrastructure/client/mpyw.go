//go:generate mockgen -source=mpyw.go -destination=mpyw_mock.go -package=client -self_package=github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client

package client

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper"
	"github.com/pkg/errors"
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
		return result, errors.Wrap(err, "failed to parse given targetURL")
	}

	q := u.Query()
	q.Add("count", fmt.Sprint(count))
	u.RawQuery = q.Encode()

	res, err := m.Client.Get(u.String())
	if err != nil {
		return result, errors.Wrap(err, "failed to fetch challenge result")
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return result, errors.Wrap(err, "failed to decode challenge result")
	}

	return result, nil
}
