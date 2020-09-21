//go:generate mockgen -source=time.go -destination=time_mock.go -package=wrapper -self_package=github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper

package wrapper

import "time"

type Timer interface {
	After(d time.Duration) <-chan time.Time
}

type timer struct{}

func NewTimer() Timer {
	return &timer{}
}

func (t *timer) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
