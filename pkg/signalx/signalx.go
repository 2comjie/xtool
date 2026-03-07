package signalx

import (
	"os"
	"os/signal"
	"syscall"
)

// Wait blocks until SIGINT or SIGTERM is received, then calls fn (if non-nil).
func Wait(fn func()) {
	WaitFor(fn, syscall.SIGINT, syscall.SIGTERM)
}

// WaitFor blocks until one of the given signals is received, then calls fn (if non-nil).
func WaitFor(fn func(), sigs ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)
	<-ch
	signal.Stop(ch)
	if fn != nil {
		fn()
	}
}

// Go spawns a goroutine that waits for SIGINT/SIGTERM, then calls fn.
func Go(fn func()) {
	go Wait(fn)
}
