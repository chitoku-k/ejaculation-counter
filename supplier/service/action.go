package service

type actionResult struct {
	Event Event
	Index int
}

type Action interface {
	Name() string
	Target(message Message) bool
	Event(message Message) (Event, int, error)
}
