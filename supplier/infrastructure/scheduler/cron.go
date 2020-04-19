package scheduler

import (
	"time"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

type scheduler struct {
	cron        *cron.Cron
	ch          chan service.Event
	Environment config.Environment
}

func New(environment config.Environment) (service.Scheduler, error) {
	s := scheduler{
		cron:        cron.New(),
		ch:          make(chan service.Event),
		Environment: environment,
	}

	_, err := s.cron.AddFunc("00 00 * * *", s.handle)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register schedule")
	}

	return &s, nil
}

func (s *scheduler) Start() <-chan service.Event {
	s.cron.Start()
	return s.ch
}

func (s *scheduler) handle() {
	s.ch <- &service.UpdateEvent{
		Date:   time.Now().Format(time.RFC3339),
		UserID: s.Environment.Mastodon.UserID,
	}
}
