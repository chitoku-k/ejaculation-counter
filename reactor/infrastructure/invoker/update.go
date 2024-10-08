package invoker

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/mattn/go-mastodon"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	DisplayNameRegex = regexp.MustCompile(`(.*)（昨日: (\d+) \/ 今日: (\d+)）`)
	UpdatesTotal     = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "updates_total",
		Help:      "Total number of updates through API.",
	})
	UpdatesErrorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ejaculation_counter",
		Name:      "updates_error_total",
		Help:      "Total number of errors triggered when updating through API.",
	}, []string{"type"})
)

type update struct {
	Client *mastodon.Client
	DB     client.DB
	UserID int64
}

type summary struct {
	Name      string
	Yesterday int
	Today     int
}

func NewUpdate(
	client *mastodon.Client,
	db client.DB,
	userID int64,
) service.Update {
	return &update{
		Client: client,
		DB:     db,
		UserID: userID,
	}
}

func parse(account mastodon.Account) summary {
	matches := DisplayNameRegex.FindStringSubmatch(account.DisplayName)
	if matches == nil || matches[1] == "" {
		return summary{
			Name:      account.DisplayName,
			Yesterday: 0,
			Today:     0,
		}
	}

	name := matches[1]
	yesterday, _ := strconv.Atoi(matches[2])
	today, _ := strconv.Atoi(matches[3])

	return summary{
		Name:      name,
		Yesterday: yesterday,
		Today:     today,
	}
}

func message(date time.Time, summary summary) string {
	status := date.Format(time.DateOnly + " ")

	if summary.Yesterday == summary.Today {
		status += "も"
	} else {
		status += "は"
	}

	if summary.Today > 0 {
		return fmt.Sprintf("%s %d 回ぴゅっぴゅしました…", status, summary.Today)
	}

	return fmt.Sprintf("%sぴゅっぴゅしませんでした…", status)
}

func (u *update) Do(ctx context.Context, event service.UpdateEvent) error {
	user, err := u.Client.GetAccountCurrentUser(ctx)
	if err != nil {
		UpdatesErrorTotal.WithLabelValues("get").Inc()
		return fmt.Errorf("failed to get current user for updating: %w", err)
	}

	summary := parse(*user)
	name := fmt.Sprintf(
		"%s（昨日: %d / 今日: %d）",
		summary.Name,
		summary.Today,
		0,
	)

	_, err = u.Client.AccountUpdate(ctx, &mastodon.Profile{
		DisplayName: &name,
	})
	if err != nil {
		UpdatesErrorTotal.WithLabelValues("update").Inc()
		return fmt.Errorf("failed to update current user: %w", err)
	}

	date := time.Date(event.Year, time.Month(event.Month), event.Day, 0, 0, 0, 0, time.Local)
	yesterday := date.AddDate(0, 0, -1)
	_, err = u.Client.PostStatus(ctx, &mastodon.Toot{
		Status:     message(yesterday, summary),
		Visibility: "private",
	})
	if err != nil {
		UpdatesErrorTotal.WithLabelValues("toot").Inc()
		return fmt.Errorf("failed to send update: %w", err)
	}

	err = u.DB.UpdateCount(ctx, u.UserID, yesterday, summary.Today)
	if err != nil {
		UpdatesErrorTotal.WithLabelValues("db").Inc()
		return fmt.Errorf("failed to update DB: %w", err)
	}

	UpdatesTotal.Inc()
	return nil
}
