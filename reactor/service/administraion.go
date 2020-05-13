package service

import "context"

type Administration interface {
	Do(ctx context.Context, event AdministrationEvent) error
}
