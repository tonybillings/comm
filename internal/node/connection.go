package node

import (
	"tonysoft.com/comm/internal/socket"
	"tonysoft.com/comm/pkg/client"
)

type ConnectionType int

const (
	Caller ConnectionType = iota
	Callee
)

type Connection struct {
	socket.DefaultConnection

	connectionType ConnectionType

	// Will be nil if this node is the callee
	callerConn client.Client

	// Will be nil if this node is the caller
	calleeConn socket.Connection
}

func NewConnection(remoteAddress string, idleTimeoutMs int64,
	closeHandler func(socket.ConnectionID) error) *Connection {
	conn := &Connection{}
	conn.Configure(remoteAddress, idleTimeoutMs, closeHandler)
	return conn
}

func (c *Connection) Read(buffer []byte) (int, error) {
	switch c.connectionType {
	case Caller:
		return c.callerConn.Read(buffer)
	case Callee:
		return c.calleeConn.Read(buffer)
	}
	return 0, nil
}

func (c *Connection) Write(data []byte) (int, error) {
	switch c.connectionType {
	case Caller:
		return c.callerConn.Write(data)
	case Callee:
		return c.calleeConn.Write(data)
	}
	return 0, nil
}
