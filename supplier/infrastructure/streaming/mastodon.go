package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/gorilla/websocket"
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
	ch          chan service.Status
	conn        wrapper.Conn
	Environment config.Environment
	Client      *mast.Client
	Dialer      wrapper.Dialer
	Timer       wrapper.Timer
}

func NewMastodon(
	environment config.Environment,
	dialer wrapper.Dialer,
	timer wrapper.Timer,
) service.Streaming {
	return &mastodon{
		ch:          make(chan service.Status),
		Environment: environment,
		Client: mast.NewClient(&mast.Config{
			Server:      environment.Mastodon.ServerURL,
			AccessToken: environment.Mastodon.AccessToken,
		}),
		Dialer: dialer,
		Timer:  timer,
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

func (m *mastodon) Statuses() <-chan service.Status {
	return m.ch
}

func (m *mastodon) Close(exit bool) error {
	if exit {
		close(m.ch)
		m.ch = nil
	}
	if m.conn != nil {
		m.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Shutdown"))
		return m.conn.Close()
	}
	return nil
}

func (m *mastodon) reconnect(ctx context.Context, current time.Duration, err error) (time.Duration, error) {
	var reconnect time.Duration
	select {
	case <-ctx.Done():
		return reconnect, ctx.Err()

	case m.ch <- service.Error{Err: err}:
		break
	}

	reconnect = time.Duration(
		math.Min(
			math.Max(
				float64(current*2),
				float64(ReconnectInitial),
			),
			float64(ReconnectMax),
		),
	)
	StreamingRetryTotal.Inc()

	select {
	case <-ctx.Done():
		return reconnect, ctx.Err()

	case m.ch <- service.Reconnection{In: reconnect}:
		break
	}

	select {
	case <-ctx.Done():
		return reconnect, ctx.Err()

	case <-m.Timer.After(reconnect):
		return reconnect, nil
	}
}

func (m *mastodon) disconnect(ctx context.Context, err error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()

	case m.ch <- service.Disconnection{Err: err}:
		err = m.Close(false)
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()

			case m.ch <- service.Error{Err: err}:
				m.conn = nil
				return nil
			}
		}
	}

	return nil
}

func (m *mastodon) Run(ctx context.Context) error {
	reconnect := ReconnectNone

	params := url.Values{}
	params.Set("access_token", m.Client.Config.AccessToken)
	params.Set("stream", m.Environment.Mastodon.Stream)

	u, err := url.Parse(m.Client.Config.Server)
	if err != nil {
		return fmt.Errorf("failed to parse server URL: %w", err)
	}
	u.Scheme = strings.ReplaceAll(u.Scheme, "http", "ws")
	u.Path = path.Join(u.Path, "/api/v1/streaming")
	u.RawQuery = params.Encode()

	for {
		logrus.Debugf("Connecting to streaming...")

		var res *http.Response
		m.conn, res, err = m.Dialer.DialContext(ctx, u.String(), nil)
		if err != nil {
			if res != nil {
				err = fmt.Errorf("failed to connect: %v: %w", res.Status, err)
			}

			reconnect, err = m.reconnect(ctx, reconnect, err)
			if err != nil {
				return err
			}
			continue
		}

		reconnect = ReconnectNone
		server := res.Header.Get(ServerHeader)
		if server != "" {
			StreamingMessageTotal.WithLabelValues(server)
			select {
			case <-ctx.Done():
				return ctx.Err()

			case m.ch <- service.Connection{Server: server}:
				break
			}
		}

		for {
			var stream mast.Stream
			err = m.conn.ReadJSON(&stream)
			if err != nil {
				err = m.disconnect(ctx, err)
				if err != nil {
					return err
				}
				break
			}

			var status mast.Status
			switch stream.Event {
			case "update":
				err = json.NewDecoder(strings.NewReader(stream.Payload.(string))).Decode(&status)

			case "conversation":
				var conversation mast.Conversation
				err = json.NewDecoder(strings.NewReader(stream.Payload.(string))).Decode(&conversation)
				if conversation.LastStatus != nil {
					status = *conversation.LastStatus
				}

			default:
				continue
			}

			if err != nil {
				select {
				case <-ctx.Done():
					return ctx.Err()

				case m.ch <- service.Error{Err: err}:
					continue
				}
			}

			message := service.Message{
				ID: string(status.ID),
				Account: service.Account{
					ID:          string(status.Account.ID),
					Acct:        status.Account.Acct,
					DisplayName: status.Account.DisplayName,
					Username:    status.Account.Username,
				},
				CreatedAt:  status.CreatedAt,
				Content:    convertContent(status.Content),
				Emojis:     convertEmojis(status.Emojis),
				IsReblog:   status.Reblog != nil,
				Tags:       convertTags(status.Tags),
				Visibility: status.Visibility,
			}

			if status.InReplyToID != nil {
				message.InReplyToID = fmt.Sprint(status.InReplyToID)
			}

			StreamingMessageTotal.WithLabelValues(server).Inc()
			select {
			case <-ctx.Done():
				return ctx.Err()

			case m.ch <- message:
				break
			}
		}
	}
}
