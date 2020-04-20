package service

import "time"

type Count struct {
	UserID int64     `db:"user"`
	Date   time.Time `db:"date"`
	Count  int       `db:"count"`
}

type User struct {
	ID         int64  `db:"id"`
	ScreenName string `db:"screen_name"`
}
