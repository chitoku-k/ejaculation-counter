package client

import (
	"github.com/mattn/go-mastodon"
)

func NewMastodon(server, accessToken string) *mastodon.Client {
	return mastodon.NewClient(&mastodon.Config{
		Server:      server,
		AccessToken: accessToken,
	})
}
