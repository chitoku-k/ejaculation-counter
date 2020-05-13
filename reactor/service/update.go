package service

import "context"

type Update interface {
	Do(ctx context.Context, event UpdateEvent) error
}
