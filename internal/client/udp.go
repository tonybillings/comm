package client

import (
	"fmt"
	"net"
	"strconv"
	"time"
	"tonysoft.com/comm/internal/socket"
	"tonysoft.com/comm/pkg/comerr"
)

type UdpClient struct {
	BaseClient
	conn          *net.UDPConn
	readTimeoutUs int
}

func (c *UdpClient) Start() error {
	if c.IsConnected() {
		return comerr.ErrClientAlreadyConnected
	}

	cfg := c.Config()

	c.readTimeoutUs = cfg.ReadTimeoutUs

	remoteAddr := cfg.RemoteAddress + ":" + strconv.FormatUint(uint64(cfg.RemotePort), 10)
	addr, err := net.ResolveUDPAddr("udp", remoteAddr)
	if err != nil {
		return err
	}

	dialer := net.Dialer{Timeout: time.Duration(cfg.ConnectTimeoutSec) * time.Second}
	udpConn, err := dialer.Dial("udp4", addr.String())
	if err != nil {
		return err
	}

	c.conn = udpConn.(*net.UDPConn)
	c.SetIsConnected(true)

	return nil
}

func (c *UdpClient) Stop() error {
	defer c.SetIsConnected(false)
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

func (c *UdpClient) Read(buffer []byte) (int, error) {
	if c.conn == nil {
		return -1, net.ErrClosed
	}

	err := c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.readTimeoutUs) * time.Microsecond))
	if err != nil {
		return -1, fmt.Errorf("%w : %v", comerr.ErrSetReadTimeout, err)
	}

	count, err := c.conn.Read(buffer)
	err = socket.SinkReadWriteError(err)
	if err != nil {
		_ = c.Stop()
	}
	return count, err
}

func (c *UdpClient) Write(data []byte) (int, error) {
	if c.conn == nil {
		return -1, net.ErrClosed
	}

	count, err := c.conn.Write(data)
	err = socket.SinkReadWriteError(err)
	if err != nil {
		_ = c.Stop()
	}
	return count, err
}
