package service

import (
	"io"
)

type Event interface {
	Name() string
}

type ReplyEvent struct {
	InReplyToID string
	Acct        string
	Body        io.ReadCloser
	Visibility  string
}

func (ReplyEvent) Name() string {
	return "events.reply"
}

type ReplyErrorEvent struct {
	InReplyToID string
	Acct        string
	ActionName  string
	Visibility  string
}

func (ReplyErrorEvent) Name() string {
	return "events.reply_error"
}

type UpdateEvent struct {
	Year  int
	Month int
	Day   int
}

func (UpdateEvent) Name() string {
	return "events.update"
}

type IncrementEvent struct {
	Year  int
	Month int
	Day   int
}

func (IncrementEvent) Name() string {
	return "events.increment"
}

type AdministrationEvent struct {
	InReplyToID string
	Acct        string
	Type        string
	Command     string
	Visibility  string
}

func (AdministrationEvent) Name() string {
	return "events.administration"
}
