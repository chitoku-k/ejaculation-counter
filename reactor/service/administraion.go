package service

type Administration interface {
	Do(event AdministrationEvent) error
}
