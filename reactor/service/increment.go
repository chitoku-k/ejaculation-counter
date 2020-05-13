package service

import "context"

type Increment interface {
	Do(ctx context.Context, event IncrementEvent) error
}
