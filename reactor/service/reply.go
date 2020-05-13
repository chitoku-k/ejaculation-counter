package service

import "context"

type Reply interface {
	Send(ctx context.Context, event ReplyEvent) error
}
