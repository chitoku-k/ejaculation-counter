package service

import "time"

type Status interface {
	status()
}

type Message struct {
	ID          string
	Account     Account
	CreatedAt   time.Time
	Content     string
	Emojis      []Emoji
	InReplyToID string
	IsReblog    bool
	Tags        []Tag
	Visibility  string
}

type Account struct {
	ID          string
	Acct        string
	DisplayName string
	Username    string
}

type Emoji struct {
	Shortcode string
}

type Tag struct {
	Name string
}

func (m Message) status() {}

type Connection struct {
	Server string
}

func (c Connection) status() {}

type Disconnection struct {
	Err error
}

func (d Disconnection) status() {}

type Reconnection struct {
	In time.Duration
}

func (r Reconnection) status() {}

type Error struct {
	Err error
}

func (e Error) status() {}
