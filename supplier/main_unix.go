//go:build unix

package main

import (
	"golang.org/x/sys/unix"
)

func init() {
	signals = append(signals, unix.SIGTERM)
}
