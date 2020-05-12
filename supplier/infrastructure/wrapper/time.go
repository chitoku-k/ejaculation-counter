//go:generate mockgen -source=time.go -destination=time_mock.go -package=wrapper -self_package=github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper

package wrapper

import "time"

type Ticker interface {
	Tick(d time.Duration) <-chan time.Time
}

type ticker struct{}

func NewTicker() Ticker {
	return &ticker{}
}

func (t *ticker) Tick(d time.Duration) <-chan time.Time {
	return time.Tick(d)
}
