package service

import (
	"strconv"
	"time"
)

type Tick struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

func (t Tick) status() {}

func (t Tick) Name() string {
	return "events.tick"
}

func (t Tick) HashCode() int64 {
	hash := 7
	hash = 31*hash + t.Year
	hash = 31*hash + t.Month
	hash = 31*hash + t.Day
	return int64(hash)
}

type Message struct {
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

func (m Message) status() {}

func (m Message) Name() string {
	return "events.message"
}

func (m Message) HashCode() int64 {
	id, _ := strconv.ParseInt(m.ID, 10, 64)
	accountID, _ := strconv.ParseInt(m.Account.ID, 10, 64)

	var hash int64
	hash = 7
	hash = 31*hash + id
	hash = 31*hash + accountID
	return hash
}

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
