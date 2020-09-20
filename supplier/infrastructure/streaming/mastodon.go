package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	mast "github.com/mattn/go-mastodon"
	"github.com/microcosm-cc/bluemonday"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

var (
	StreamingMessageTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "streaming_message_total",
		Help:      "Total number of messages from streaming.",
	}, []string{"server"})
	StreamingRetryTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "streaming_retry_total",
		Help:      "Total number of retry for streaming.",
	})
)

const (
	ReconnectNone    = 0 * time.Second
	ReconnectInitial = 5 * time.Second
	ReconnectMax     = 320 * time.Second
	ServerHeader     = "X-Served-By"
)

type mastodon struct {
	Environment config.Environment
	Client      *mast.Client
	Dialer      wrapper.Dialer
	Ticker      wrapper.Ticker
}

func NewMastodon(
	environment config.Environment,
	dialer wrapper.Dialer,
	ticker wrapper.Ticker,
) service.Streaming {
	return &mastodon{
		Environment: environment,
		Client: mast.NewClient(&mast.Config{
			Server:      environment.Mastodon.ServerURL,
			AccessToken: environment.Mastodon.AccessToken,
		}),
		Dialer: dialer,
		Ticker: ticker,
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

func (m *mastodon) Run(ctx context.Context) (<-chan service.Status, error) {
	reconnect := ReconnectNone
	ch := make(chan service.Status)

	params := url.Values{}
	params.Set("access_token", m.Client.Config.AccessToken)
	params.Set("stream", m.Environment.Mastodon.Stream)

	u, err := url.Parse(m.Client.Config.Server)
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("failed to parse server URL: %w", err)
	}
	u.Scheme = strings.ReplaceAll(u.Scheme, "http", "ws")
	u.Path = path.Join(u.Path, "/api/v1/streaming")
	u.RawQuery = params.Encode()

	go func() {
		defer close(ch)

		for {
			if ctx.Err() != nil {
				return
			}

			conn, res, err := m.Dialer.DialContext(ctx, u.String(), nil)
			if err != nil {
				if res != nil {
					ch <- service.Error{
						Err: fmt.Errorf("failed to connect: %v: %w", res.Status, err),
					}
				} else {
					ch <- service.Error{
						Err: err,
					}
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
				ch <- service.Reconnection{
					In: reconnect,
				}
				StreamingRetryTotal.Inc()
				// TODO: The usage of time.Tick simply leaks
				<-m.Ticker.Tick(reconnect)
				continue
			}

			defer func() {
				if conn != nil {
					conn.Close()
				}
			}()

			reconnect = ReconnectNone
			server := res.Header.Get(ServerHeader)
			if server != "" {
				logrus.Infof("Connected to Mastodon: %v", server)
				StreamingMessageTotal.WithLabelValues(server)
			}

			for {
				if ctx.Err() != nil {
					return
				}

				var stream mast.Stream
				err = conn.ReadJSON(&stream)
				if err != nil {
					ch <- service.Error{
						Err: err,
					}
					err = conn.Close()
					if err != nil {
						ch <- service.Error{
							Err: err,
						}
					}
					conn = nil
					break
				}

				var status mast.Status
				switch stream.Event {
				case "update":
					err = json.NewDecoder(strings.NewReader(stream.Payload.(string))).Decode(&status)

				case "conversation":
					var conversation mast.Conversation
					err = json.NewDecoder(strings.NewReader(stream.Payload.(string))).Decode(&conversation)
					status = *conversation.LastStatus

				default:
					continue
				}

				if err != nil {
					ch <- service.Error{
						Err: err,
					}
					continue
				}

				StreamingMessageTotal.WithLabelValues(server).Inc()
				ch <- service.Message{
					ID: string(status.ID),
					Account: service.Account{
						ID:          string(status.Account.ID),
						Acct:        status.Account.Acct,
						DisplayName: status.Account.DisplayName,
						Username:    status.Account.Username,
					},
					CreatedAt:   status.CreatedAt,
					Content:     convertContent(status.Content),
					Emojis:      convertEmojis(status.Emojis),
					InReplyToID: fmt.Sprint(status.InReplyToID),
					IsReblog:    status.Reblog != nil,
					Tags:        convertTags(status.Tags),
					Visibility:  status.Visibility,
				}
			}
		}
	}()

	return ch, nil
}
