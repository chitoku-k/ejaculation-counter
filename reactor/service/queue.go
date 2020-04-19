package service

type QueueReader interface {
	Consume() (<-chan Event, error)
}
