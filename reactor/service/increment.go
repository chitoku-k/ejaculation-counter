package service

type Increment interface {
	Do(event IncrementEvent) error
}
