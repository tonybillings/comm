package server

import (
	"net"
	"time"
	"tonysoft.com/comm/internal/socket"
)

type Connection struct {
	socket.DefaultConnection

	server ReadWriter

	tcpConn    *net.TCPConn
	udpConn    *UdpConn
	rfcommConn int
}

func (c *Connection) Read(buffer []byte) (int, error) {
	return c.server.read(c, buffer)
}

func (c *Connection) Write(data []byte) (int, error) {
	return c.server.write(c, data)
}

func (c *Connection) ConfigureTCP(server ReadWriter, conn *net.TCPConn, idleTimeoutMs int64,
	closeHandler func(socket.ConnectionID) error) {
	c.DefaultConnection.Configure(conn.RemoteAddr().String(), idleTimeoutMs, closeHandler)
	c.server = server
	c.tcpConn = conn
}

func (c *Connection) ConfigureUDP(server ReadWriter, conn *UdpConn) {
	c.DefaultConnection.Configure(conn.RemoteAddr().String(), -1, nil)
	c.SetDisconnectTime(time.Now().UTC())
	c.SetIsConnected(false)
	c.server = server
	c.udpConn = conn
}

func (c *Connection) ConfigureRFCOMM(server ReadWriter, conn int, remoteAddress string, idleTimeoutMs int64,
	closeHandler func(socket.ConnectionID) error) {
	c.DefaultConnection.Configure(remoteAddress, idleTimeoutMs, closeHandler)
	c.server = server
	c.rfcommConn = conn
}
