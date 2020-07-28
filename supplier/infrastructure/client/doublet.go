//go:generate mockgen -source=doublet.go -destination=doublet_mock.go -package=client -self_package=github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client

package client

import (
	"encoding/json"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper"
	"github.com/pkg/errors"
)

type doublet struct {
	Client wrapper.HttpClient
}

type Doublet interface {
	Do(targetURL string) (DoubletResult, error)
}

type DoubletResult []string

func NewDoublet(client wrapper.HttpClient) Doublet {
	return &doublet{
		Client: client,
	}
}

func (t *doublet) Do(targetURL string) (DoubletResult, error) {
	res, err := t.Client.Get(targetURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch challenge result")
	}
	defer res.Body.Close()

	var result DoubletResult
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode challenge result")
	}

	return result, nil
}
