//go:generate go tool mockgen -source=queue.go -destination=queue_mock.go -package=service -self_package=github.com/chitoku-k/ejaculation-counter/supplier/service

package service

import "context"

type QueueWriter interface {
	Publish(ctx context.Context, packet Packet) error
	Close() error
}
