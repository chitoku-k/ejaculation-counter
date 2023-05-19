package main

import (
	"syscall"

	"golang.org/x/sys/windows"
)

func init() {
	// The signals CTRL_CLOSE_EVENT, CTRL_LOGOFF_EVENT, and CTRL_SHUTDOWN_EVENT are treated as SIGTERM in runtime package.
	// See also: https://learn.microsoft.com/en-us/windows/console/setconsolectrlhandler
	signals = append(signals, syscall.Signal(windows.SIGTERM))
}
