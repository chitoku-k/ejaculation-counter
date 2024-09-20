package invoker

import (
	"context"
	"fmt"
	"time"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/mattn/go-mastodon"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	IncrementTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "increment_total",
		Help:      "Total number of increment through API.",
	})
	IncrementErrorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "increment_error_total",
		Help:      "Total number of errors triggered when incrementing through API.",
	}, []string{"type"})
)

type increment struct {
	Client *mastodon.Client
	DB     client.DB
	UserID int64
}

func NewIncrement(
	client *mastodon.Client,
	db client.DB,
	userID int64,
) service.Increment {
	return &increment{
		Client: client,
		DB:     db,
		UserID: userID,
	}
}

func (i *increment) Do(ctx context.Context, event service.IncrementEvent) error {
	user, err := i.Client.GetAccountCurrentUser(ctx)
	if err != nil {
		IncrementErrorTotal.WithLabelValues("get").Inc()
		return fmt.Errorf("failed to get current user for updating: %w", err)
	}

	summary := parse(*user)
	name := fmt.Sprintf(
		"%s（昨日: %d / 今日: %d）",
		summary.Name,
		summary.Yesterday,
		summary.Today+1,
	)

	_, err = i.Client.AccountUpdate(ctx, &mastodon.Profile{
		DisplayName: &name,
	})
	if err != nil {
		IncrementErrorTotal.WithLabelValues("update").Inc()
		return fmt.Errorf("failed to update current user: %w", err)
	}

	err = i.DB.UpdateCount(ctx, i.UserID, time.Now(), summary.Today+1)
	if err != nil {
		IncrementErrorTotal.WithLabelValues("db").Inc()
		return fmt.Errorf("failed to update DB: %w", err)
	}

	IncrementTotal.Inc()
	return nil
}
