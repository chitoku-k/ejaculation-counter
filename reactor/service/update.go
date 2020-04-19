package service

type Update interface {
	Do(event UpdateEvent) error
}
