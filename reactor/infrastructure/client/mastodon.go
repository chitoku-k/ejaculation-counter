package client

import (
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/mattn/go-mastodon"
)

func NewMastodon(environment config.Environment) *mastodon.Client {
	return mastodon.NewClient(&mastodon.Config{
		Server:      environment.Mastodon.ServerURL,
		AccessToken: environment.Mastodon.AccessToken,
	})
}
