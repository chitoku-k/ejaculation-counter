package service

type Packet interface {
	Name() string
	HashCode() int64
}
