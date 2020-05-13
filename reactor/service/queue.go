package service

import "context"

type QueueReader interface {
	Consume(ctx context.Context) (<-chan Event, error)
}
