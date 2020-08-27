//go:generate mockgen -source=doublet.go -destination=doublet_mock.go -package=client -self_package=github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/client

package client

import (
	"encoding/json"
	"fmt"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper"
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
