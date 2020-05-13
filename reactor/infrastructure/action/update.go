package actions

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/client"
	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	"github.com/mattn/go-mastodon"
	"github.com/pkg/errors"
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
	Environment config.Environment
	Client      *mastodon.Client
	DB          client.DB
}

type summary struct {
	Name      string
	Yesterday int
	Today     int
}

func NewUpdate(
	environment config.Environment,
	client *mastodon.Client,
	db client.DB,
) service.Update {
	return &update{
		Environment: environment,
		Client:      client,
		DB:          db,
	}
}

func parse(account mastodon.Account) summary {
	matches := DisplayNameRegex.FindStringSubmatch(account.DisplayName)

	name := matches[1]
	if name == "" {
		name = account.DisplayName
	}

	yesterday, _ := strconv.Atoi(matches[2])
	today, _ := strconv.Atoi(matches[3])

	return summary{
		Name:      name,
		Yesterday: yesterday,
		Today:     today,
	}
}

func message(date time.Time, summary summary) string {
	status := date.Format("2006/01/02 ")

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
		return errors.Wrap(err, "failed to get current user for updating")
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
		return errors.Wrap(err, "failed to update current user")
	}

	date, err := time.Parse(time.RFC3339, event.Date)
	if err != nil {
		UpdatesErrorTotal.WithLabelValues("parse").Inc()
		return errors.Wrap(err, "failed to parse update date")
	}

	yesterday := date.AddDate(0, 0, -1)
	_, err = u.Client.PostStatus(ctx, &mastodon.Toot{
		Status:     message(yesterday, summary),
		Visibility: "private",
	})
	if err != nil {
		UpdatesErrorTotal.WithLabelValues("toot").Inc()
		return errors.Wrap(err, "failed to send update")
	}

	err = u.DB.UpdateCount(ctx, u.Environment.UserID, yesterday, summary.Today)
	if err != nil {
		UpdatesErrorTotal.WithLabelValues("db").Inc()
		return errors.Wrap(err, "failed to update DB")
	}

	UpdatesTotal.Inc()
	return nil
}
