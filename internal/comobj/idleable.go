package comobj

import (
	"sync/atomic"
	"time"
)

// Idleable Read-only interface for objects that maintain idle state
type Idleable interface {
	IsIdle() bool
}

type DefaultIdleable struct {
	lastNotIdle   atomic.Int64
	idleTimeoutMs atomic.Int64
}

func (i *DefaultIdleable) SetIdleTimeout(milliseconds int64) {
	i.idleTimeoutMs.Store(milliseconds)
	i.NotIdle()
}

func (i *DefaultIdleable) IsIdle() bool {
	idleTimeoutMs := i.idleTimeoutMs.Load()
	if idleTimeoutMs < 1 {
		return false
	}
	return time.Now().UnixMilli()-i.lastNotIdle.Load() > idleTimeoutMs
}

func (i *DefaultIdleable) NotIdle() {
	i.lastNotIdle.Store(time.Now().UnixMilli())
}
