package service

type Streaming interface {
	Run() (<-chan MessageStatus, error)
}
