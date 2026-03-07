package taskx

import "errors"

var (
	ErrPoolStopped = errors.New("pool has been stopped")
	ErrPoolFull    = errors.New("pool queue is full")
)
