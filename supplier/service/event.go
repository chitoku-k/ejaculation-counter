package service

type Event interface {
	Name() string
	HashCode() int64
}
