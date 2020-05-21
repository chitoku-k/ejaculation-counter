//go:generate mockgen -source=queue.go -destination=queue_mock.go -package=service -self_package=github.com/chitoku-k/ejaculation-counter/supplier/service

package service

type QueueWriter interface {
	Publish(event Event) error
}
