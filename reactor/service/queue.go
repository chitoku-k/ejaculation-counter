package service

import "context"

type QueueReader interface {
	Consume(ctx context.Context)
	Packets() <-chan Packet
	Close(exit bool) error
}
