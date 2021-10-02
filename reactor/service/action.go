//go:generate mockgen -source=action.go -destination=action_mock.go -package=service -self_package=github.com/chitoku-k/ejaculation-counter/reactor/service

package service

import "context"

type actionResult struct {
	Event Event
	Index int
}

type Action interface {
	Name() string
	Target(message Message) bool
	Event(ctx context.Context, message Message) (Event, int, error)
}
