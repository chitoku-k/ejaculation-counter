package service

import (
	"time"
)

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
