package service

type Scheduler interface {
	Start() <-chan Event
}
