package comobj

import (
	"sync/atomic"
)

// Runnable Read-only interface for objects that maintain running state
type Runnable interface {
	IsRunning() bool
}

type DefaultRunnable struct {
	isRunning atomic.Bool
}

func (r *DefaultRunnable) IsRunning() bool {
	return r.isRunning.Load()
}

func (r *DefaultRunnable) SetIsRunning(isRunning bool) {
	r.isRunning.Store(isRunning)
}
