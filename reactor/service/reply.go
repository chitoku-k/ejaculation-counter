package service

type Reply interface {
	Send(event ReplyEvent) error
}
