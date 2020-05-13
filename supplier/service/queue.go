package service

type QueueWriter interface {
	Publish(event Event) error
}
