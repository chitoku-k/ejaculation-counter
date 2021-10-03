package service

import (
	"time"
)

type Packet interface {
	Tag() uint64
}

func NewTick(tag uint64) Tick {
	return Tick{
		tag: tag,
	}
}

type Tick struct {
	tag uint64

	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

func (t Tick) Name() string {
	return "packets.tick"
}

func (t Tick) Tag() uint64 {
	return t.tag
}

func NewMessage(tag uint64) Message {
	return Message{
		tag: tag,
	}
}

type Message struct {
	tag uint64

	ID          string    `json:"id"`
	Account     Account   `json:"account"`
	CreatedAt   time.Time `json:"created_at"`
	Content     string    `json:"content"`
	Emojis      []Emoji   `json:"emojis"`
	InReplyToID string    `json:"in_reply_to_id"`
	IsReblog    bool      `json:"is_reblog"`
	Tags        []Tag     `json:"tags"`
	Visibility  string    `json:"visibility"`
}

type Account struct {
	ID          string `json:"id"`
	Acct        string `json:"acct"`
	DisplayName string `json:"display_name"`
	Username    string `json:"user_name"`
}

type Emoji struct {
	Shortcode string `json:"shortcode"`
}

type Tag struct {
	Name string `json:"name"`
}

func (m Message) Name() string {
	return "packets.message"
}

func (m Message) Tag() uint64 {
	return m.tag
}
