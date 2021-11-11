package service

import "context"

type QueueReader interface {
	Consume(ctx context.Context)
	Packets() <-chan Packet
	Ack(tag uint64) error
	Reject(tag uint64) error
	Close(exit bool) error
}
