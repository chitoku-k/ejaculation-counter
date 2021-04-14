package scheduler

import (
	"fmt"
	"time"

	"github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/config"
	"github.com/chitoku-k/ejaculation-counter/supplier/service"
	"github.com/robfig/cron/v3"
)

type scheduler struct {
	cron        *cron.Cron
	ch          chan service.Tick
	Environment config.Environment
}

func New(environment config.Environment) (service.Scheduler, error) {
	s := scheduler{
		cron:        cron.New(),
		ch:          make(chan service.Tick),
		Environment: environment,
	}

	_, err := s.cron.AddFunc("00 00 * * *", s.handle)
	if err != nil {
		return nil, fmt.Errorf("failed to register schedule: %w", err)
	}

	return &s, nil
}

func (s *scheduler) Start() <-chan service.Tick {
	s.cron.Start()
	return s.ch
}

func (s *scheduler) Stop() {
	defer close(s.ch)
	<-s.cron.Stop().Done()
}

func (s *scheduler) handle() {
	year, month, day := time.Now().Local().Date()
	s.ch <- service.Tick{
		Year:  year,
		Month: int(month),
		Day:   day,
	}
}
