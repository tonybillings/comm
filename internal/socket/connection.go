package socket

import (
	"io"
	"sync/atomic"
	"time"
	"tonysoft.com/comm/internal/comobj"
	"tonysoft.com/comm/pkg/comerr"
)

type ConnectionID uint64

var (
	nextConnectionId atomic.Uint64
)

func NewConnectionID() ConnectionID {
	return ConnectionID(nextConnectionId.Add(1))
}

type Connection interface {
	ID() ConnectionID
	RemoteAddress() string
	comobj.Connectable
	comobj.Idleable
	io.Reader
	io.Writer
	io.Closer
}

type DefaultConnection struct {
	id            ConnectionID
	remoteAddress string

	comobj.DefaultConnectable
	comobj.DefaultIdleable

	closeHandler func(ConnectionID) error
}

func (c *DefaultConnection) ID() ConnectionID {
	return c.id
}

func (c *DefaultConnection) RemoteAddress() string {
	return c.remoteAddress
}

func (c *DefaultConnection) Read(_ []byte) (int, error) {
	// Structs that embed DefaultConnection to implement the
	// Connection interface should shadow/override this function
	return -1, comerr.ErrNotImplemented
}

func (c *DefaultConnection) Write(_ []byte) (int, error) {
	// Structs that embed DefaultConnection to implement the
	// Connection interface should shadow/override this function
	return -1, comerr.ErrNotImplemented
}

func (c *DefaultConnection) Close() error {
	if c == nil || !c.IsConnected() {
		return nil
	}

	c.SetIsConnected(false)
	defer c.SetDisconnectTime(time.Now().UTC())

	if c.closeHandler != nil {
		return c.closeHandler(c.id)
	}

	return nil
}

func (c *DefaultConnection) Configure(remoteAddress string, idleTimeoutMs int64, closeHandler func(ConnectionID) error) {
	c.SetConnectTime(time.Now().UTC())
	c.id = NewConnectionID()
	c.remoteAddress = remoteAddress
	c.SetIdleTimeout(idleTimeoutMs)
	c.closeHandler = closeHandler
	c.SetIsConnected(true)
}
