package comobj

import (
	"sync/atomic"
	"time"
)

// Connectable Read-only interface for objects that maintain connection state
type Connectable interface {
	IsConnected() bool
	ConnectTime() *time.Time
	DisconnectTime() *time.Time
}

type DefaultConnectable struct {
	isConnected    atomic.Bool
	connectTime    atomic.Pointer[time.Time]
	disconnectTime atomic.Pointer[time.Time]
}

func (c *DefaultConnectable) IsConnected() bool {
	return c.isConnected.Load()
}

func (c *DefaultConnectable) SetIsConnected(isConnected bool) {
	c.isConnected.Store(isConnected)
}

func (c *DefaultConnectable) ConnectTime() *time.Time {
	connectTime := c.connectTime.Load()
	if connectTime == nil {
		return nil
	} else {
		connectTimeCopy := *connectTime
		return &connectTimeCopy
	}
}

func (c *DefaultConnectable) SetConnectTime(connectTime time.Time) {
	c.connectTime.Store(&connectTime)
}

func (c *DefaultConnectable) DisconnectTime() *time.Time {
	disconnectTime := c.disconnectTime.Load()
	if disconnectTime == nil {
		return nil
	} else {
		disconnectTimeCopy := *disconnectTime
		return &disconnectTimeCopy
	}
}

func (c *DefaultConnectable) SetDisconnectTime(disconnectTime time.Time) {
	c.disconnectTime.Store(&disconnectTime)
}
