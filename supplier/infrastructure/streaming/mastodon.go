package streaming

import (
	"context"
	"fmt"
	"html"
	"math"
	"time"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	mast "github.com/mattn/go-mastodon"
	"github.com/microcosm-cc/bluemonday"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

var (
	StreamingMessageTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "streaming_message_total",
		Help:      "Total number of messages from streaming.",
	})
	StreamingRetryTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "streaming_retry_total",
		Help:      "Total number of retry for streaming.",
	})
)

const (
	ReconnectInitial = 5 * time.Second
	ReconnectMax     = 320 * time.Second
)

type mastodon struct {
	ctx         context.Context
	Environment config.Environment
	Client      *mast.Client
}

func NewMastodon(ctx context.Context, environment config.Environment) service.Streaming {
	return &mastodon{
		ctx:         ctx,
		Environment: environment,
		Client: mast.NewClient(&mast.Config{
			Server:      environment.Mastodon.ServerURL,
			AccessToken: environment.Mastodon.AccessToken,
		}),
	}
}

func stripTags(s string) string {
	return bluemonday.StrictPolicy().Sanitize(s)
}

func convertContent(content string) string {
	return html.UnescapeString(stripTags(content))
}

func convertEmojis(emojis []mast.Emoji) []service.Emoji {
	var result []service.Emoji

	for _, v := range emojis {
		result = append(result, service.Emoji{
			Shortcode: v.ShortCode,
		})
	}

	return result
}

func convertTags(tags []mast.Tag) []service.Tag {
	var result []service.Tag

	for _, v := range tags {
		result = append(result, service.Tag{
			Name: v.Name,
		})
	}

	return result
}

func (m *mastodon) Run() <-chan service.MessageStatus {
	reconnect := ReconnectInitial
	ch := make(chan service.MessageStatus)
	wsc := m.Client.NewWSClient()

	go func() {
		defer close(ch)

		for {
			if m.ctx.Err() != nil {
				return
			}

			stream, err := wsc.StreamingWSUser(m.ctx)
			if err != nil {
				ch <- service.MessageStatus{
					Error: err,
				}
				reconnect = time.Duration(
					math.Min(
						math.Max(
							float64(reconnect*2),
							float64(ReconnectInitial),
						),
						float64(ReconnectMax),
					),
				)
				logrus.Infof("Reconnecting in %v...\n", reconnect)
				StreamingRetryTotal.Inc()
				<-time.Tick(reconnect)
				continue
			}

			reconnect = ReconnectInitial
			for event := range stream {
				switch e := event.(type) {
				case *mast.UpdateEvent:
					StreamingMessageTotal.Inc()
					ch <- service.MessageStatus{
						Message: service.Message{
							ID: string(e.Status.ID),
							Account: service.Account{
								ID:          string(e.Status.Account.ID),
								Acct:        e.Status.Account.Acct,
								DisplayName: e.Status.Account.DisplayName,
								Username:    e.Status.Account.Username,
							},
							CreatedAt:   e.Status.CreatedAt,
							Content:     convertContent(e.Status.Content),
							Emojis:      convertEmojis(e.Status.Emojis),
							InReplyToID: fmt.Sprint(e.Status.InReplyToID),
							IsReblog:    e.Status.Reblog != nil,
							Tags:        convertTags(e.Status.Tags),
							Visibility:  e.Status.Visibility,
						},
					}

				case *mast.ErrorEvent:
					ch <- service.MessageStatus{
						Error: e,
					}
				}
			}
		}
	}()

	return ch
}
