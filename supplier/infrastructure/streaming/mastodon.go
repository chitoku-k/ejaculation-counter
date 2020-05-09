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
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/gorilla/websocket"
	mast "github.com/mattn/go-mastodon"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
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
	ReconnectInitial = 5 * time.Second
	ReconnectMax     = 320 * time.Second
	ServerHeader     = "X-Served-Server"
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

func (m *mastodon) Run() (<-chan service.MessageStatus, error) {
	reconnect := ReconnectInitial
	ch := make(chan service.MessageStatus)

	params := url.Values{}
	params.Set("access_token", m.Client.Config.AccessToken)
	params.Set("stream", m.Environment.Mastodon.Stream)

	u, err := url.Parse(m.Client.Config.Server)
	if err != nil {
		close(ch)
		return ch, errors.Wrap(err, "failed to parse server URL")
	}
	u.Scheme = strings.ReplaceAll(u.Scheme, "http", "ws")
	u.Path = path.Join(u.Path, "/api/v1/streaming")
	u.RawQuery = params.Encode()

	go func() {
		defer close(ch)

		for {
			if m.ctx.Err() != nil {
				return
			}

			conn, res, err := websocket.DefaultDialer.DialContext(m.ctx, u.String(), nil)
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
				logrus.Infof("Reconnecting in %v...", reconnect)
				StreamingRetryTotal.Inc()
				<-time.Tick(reconnect)
				continue
			}

			reconnect = ReconnectInitial
			server := res.Header.Get(ServerHeader)
			if server != "" {
				logrus.Infof("Connected to %s", server)
			}

			for {
				var stream mast.Stream
				err = conn.ReadJSON(&stream)
				if err != nil {
					ch <- service.MessageStatus{
						Error: err,
					}
					err = conn.Close()
					if err != nil {
						ch <- service.MessageStatus{
							Error: err,
						}
					}
					break
				}

				if stream.Event != "update" {
					continue
				}

				var status mast.Status
				err = json.Unmarshal([]byte(stream.Payload.(string)), &status)
				if err != nil {
					ch <- service.MessageStatus{
						Error: err,
					}
					continue
				}

				StreamingMessageTotal.WithLabelValues(server).Inc()
				ch <- service.MessageStatus{
					Message: service.Message{
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
					},
				}
			}
		}
	}()

	return ch, nil
}
