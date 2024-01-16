package server

import (
	"context"
	"errors"
	"net"
	"syscall"
	"tonysoft.com/comm/internal/socket"
	"tonysoft.com/comm/internal/transport"
	"tonysoft.com/comm/pkg/comerr"
	"tonysoft.com/comm/pkg/stream"
)

type UdpServer struct {
	BaseServer

	listener      *net.UDPConn
	readTimeoutUs int
}

func (s *UdpServer) Start() error {
	if s.IsRunning() {
		return comerr.ErrServerAlreadyRunning
	}

	cfg := s.Config()

	s.clearConnections()
	s.readTimeoutUs = cfg.ReadTimeoutUs

	s.ConfigureErrors(cfg.ErrorChanBufferSize)

	addr, err := transport.GetTcpAddressFromHostAndPort(cfg.Address, cfg.Port)
	if err != nil {
		return err
	}

	err = s.configureListener(addr)
	if err != nil {
		return err
	}

	go s.listenForClientConnections(cfg.ReadBufferSize)
	go s.handleListenCancel()

	s.SetIsRunning(true)

	return nil
}

// CloseClient Needed to implement Server interface
func (s *UdpServer) CloseClient(_ socket.ConnectionID) error {
	return nil
}

func (s *UdpServer) close() error {
	err := s.listener.Close()
	s.listener = nil
	s.listenContext = nil
	s.listenCancelFunc = nil

	close(s.newConnChan)
	s.CloseErrors()
	s.SetIsRunning(false)

	if errors.Is(err, syscall.EINVAL) {
		return nil
	} else {
		return err
	}
}

func (s *UdpServer) configureListener(bindAddress string) error {
	s.listener = nil

	addr, err := net.ResolveUDPAddr("udp", bindAddress)
	if err != nil {
		return err
	}

	var cfg = net.ListenConfig{
		Control: socket.ControlFunc,
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.listenContext = ctx
	s.listenCancelFunc = cancel

	listener, err := cfg.ListenPacket(s.listenContext, "udp4", addr.String())
	if err != nil {
		return err
	}

	s.listener = listener.(*net.UDPConn)

	return nil
}

func (s *UdpServer) listenForClientConnections(readBufferSize int) {
	buffer := make([]byte, readBufferSize)

	for {
		if !s.IsRunning() || s.listener == nil {
			return
		}

		count, remoteAddr, err := s.listener.ReadFrom(buffer)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			s.SendError(err)
			continue
		}

		if count < 1 {
			continue
		}

		data := make([]byte, count)
		copy(data, buffer[:count])

		udpConn := &UdpConn{}
		udpConn.DataReader = *stream.NewDataReader(data)
		udpConn.remoteAddress = remoteAddr

		conn := &Connection{}
		conn.ConfigureUDP(s, udpConn)

		select {
		case s.newConnChan <- conn:
			break
		default:
		}
	}
}

func (s *UdpServer) read(conn *Connection, buffer []byte) (int, error) {
	if conn == nil || conn.udpConn == nil {
		return -1, net.ErrClosed
	}

	count, err := conn.udpConn.Read(buffer)
	err = socket.SinkReadWriteError(err)
	return count, err
}

func (s *UdpServer) write(conn *Connection, data []byte) (int, error) {
	if conn == nil || conn.udpConn == nil || s.listener == nil {
		return -1, net.ErrClosed
	}

	count, err := s.listener.WriteTo(data, conn.udpConn.RemoteAddr())
	err = socket.SinkReadWriteError(err)
	return count, err
}

func (s *UdpServer) handleListenCancel() {
	for {
		select {
		case <-s.listenContext.Done():
			err := s.close()
			if err != nil {
				s.SendError(err)
			}
			return
		}
	}
}

/*******************************************************************************
 CONNECTION
*******************************************************************************/

type UdpConn struct {
	stream.DataReader
	remoteAddress net.Addr
}

func (c *UdpConn) RemoteAddr() net.Addr {
	return c.remoteAddress
}
