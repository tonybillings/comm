//go:build linux

package client

import (
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"syscall"
	"time"
	"tonysoft.com/comm/internal/rfcomm"
	"tonysoft.com/comm/internal/socket"
	"tonysoft.com/comm/pkg/comerr"
)

type RfcommClient struct {
	BaseClient
	conn          int
	readTimeoutUs int
}

func (c *RfcommClient) Start() error {
	c.conn = -1

	cfg := c.Config()

	c.readTimeoutUs = cfg.ReadTimeoutUs

	sock, err := unix.Socket(syscall.AF_BLUETOOTH, syscall.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		return err
	}

	sa, err := rfcomm.MacStringToByteArray(cfg.RemoteAddress)
	if err != nil {
		_ = unix.Close(sock)
		return fmt.Errorf("%w : %v", comerr.ErrParseMacAddress, err)
	}

	sockInfo := &unix.SockaddrRFCOMM{
		Addr:    sa,
		Channel: uint8(cfg.RemotePort),
	}

	rfcomm.RestartBluetooth()

	err = unix.Connect(sock, sockInfo)

	// See https://man7.org/linux/man-pages/man2/connect.2.html for the rationale behind this.
	switch err {
	case unix.EINPROGRESS, unix.EAGAIN, unix.EINTR:
		abortChan := make(chan bool)
		errChan := make(chan error)

		go func() {
			for {
				select {
				case <-abortChan:
					return
				case <-time.After(time.Second):
					fdset := unix.FdSet{}
					fdset.Set(sock)
					tv := unix.Timeval{Sec: int64(cfg.ConnectTimeoutSec), Usec: 0}

					n, selErr := unix.Select(sock+1, nil, &fdset, nil, &tv)

					if selErr == nil && n > 0 {
						optStr, getSockOptStrErr := unix.GetsockoptString(sock, unix.SOL_SOCKET, unix.SO_ERROR)

						if getSockOptStrErr == unix.EINTR {
							continue
						}

						optStrBytes := []byte(optStr)
						connected := len(optStrBytes) > 0
						for _, b := range optStrBytes {
							if b != 0 {
								connected = false
							}
						}
						connected = connected && err == nil

						if connected {
							errChan <- nil
							return
						} else {
							if err != nil {
								errChan <- err
							} else {
								errChan <- comerr.ErrConnectAborted
							}
							return
						}
					} else if err == unix.EINTR {
						continue
					} else if err != nil {
						errChan <- err
						return
					}
				}
			}
		}()

		select {
		case e := <-errChan:
			if e != nil {
				_ = unix.Close(sock)
				return e
			} else {
				break
			}
		case <-time.After(time.Duration(cfg.ConnectTimeoutSec) * time.Second):
			abortChan <- true
			_ = unix.Close(sock)
			return comerr.ErrConnectTimeout
		}
	case nil:
		break
	default:
		_ = unix.Close(sock)
		return err
	}

	c.conn = sock
	c.SetIsConnected(true)

	return c.setConnectionOptions()
}

func (c *RfcommClient) Stop() error {
	defer c.SetIsConnected(false)

	if c.conn == -1 {
		return nil
	}

	errChan := make(chan error)
	go func() {
		sock := c.conn
		c.conn = -1
		errChan <- unix.Close(sock)
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(5 * time.Second):
		return comerr.ErrDisconnectTimeout
	}
}

func (c *RfcommClient) Read(buffer []byte) (int, error) {
	if c.conn < 0 {
		return -1, net.ErrClosed
	}

	count, err := unix.Read(c.conn, buffer)
	err = socket.SinkReadWriteError(err)
	if err != nil {
		_ = c.Stop()
	}

	return count, err
}

func (c *RfcommClient) Write(data []byte) (int, error) {
	if c.conn < 0 {
		return -1, net.ErrClosed
	}

	count, err := unix.Write(c.conn, data)
	err = socket.SinkReadWriteError(err)
	if err != nil {
		_ = c.Stop()
	}

	return count, err
}

func (c *RfcommClient) setConnectionOptions() error {
	if c.conn < 0 {
		return net.ErrClosed
	}

	// The microsecond part cannot exceed domain limit
	usRemainder := c.readTimeoutUs % 1000000
	sec := c.readTimeoutUs / 1000000

	tv := unix.Timeval{Sec: int64(sec), Usec: int64(usRemainder)}
	err := unix.SetsockoptTimeval(c.conn, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &tv)
	if err != nil {
		_ = c.Stop()
		return fmt.Errorf("%w : %v", comerr.ErrSetReadTimeout, err)
	}

	err = unix.SetNonblock(c.conn, true)
	if err != nil {
		_ = c.Stop()
		return fmt.Errorf("%w : %v", comerr.ErrSetNonBlockingMode, err)
	}

	closeLingerTime := unix.Linger{Onoff: 1, Linger: 0}
	err = unix.SetsockoptLinger(c.conn, unix.SOL_SOCKET, unix.SO_LINGER, &closeLingerTime)
	if err != nil {
		_ = c.Stop()
		return fmt.Errorf("%w : %v", comerr.ErrSetLingerTimeout, err)
	}

	return nil
}
