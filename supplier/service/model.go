package service

import "time"

type MessageStatus struct {
	Message Message
	Error   error
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
