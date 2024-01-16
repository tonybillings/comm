package server

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"syscall"
	"tonysoft.com/comm/internal/rfcomm"
	"tonysoft.com/comm/internal/socket"
	"tonysoft.com/comm/pkg/comerr"
)

type RfcommServer struct {
	BaseServer

	listener         int // socket file descriptor
	listenContext    context.Context
	listenCancelFunc context.CancelFunc
}

func (s *RfcommServer) Start() error {
	if s.IsRunning() {
		return comerr.ErrServerAlreadyRunning
	}

	cfg := s.Config()

	s.clearConnections()
	s.ConfigureErrors(cfg.ErrorChanBufferSize)

	rfcomm.BecomeDiscoverable()

	err := s.configureListener(cfg.Port, cfg.ClientConnectionLimit)
	if err != nil {
		return err
	}

	go s.listenForClientConnections()
	go s.handleListenCancel()

	s.SetIsRunning(true)

	return nil
}

func (s *RfcommServer) Stop() {
	if !s.IsRunning() {
		return
	}

	if s.listenCancelFunc != nil {
		s.listenCancelFunc()
	}
}

func (s *RfcommServer) CloseClient(id socket.ConnectionID) error {
	conn, ok := s.connections.Load(id)
	if !ok {
		return nil
	}

	s.connections.Delete(id)
	return unix.Close(conn.(*Connection).rfcommConn)
}

func (s *RfcommServer) close() error {
	err := unix.Close(s.listener)
	s.listener = 0
	s.listenContext = nil
	s.listenCancelFunc = nil

	s.connections.Range(func(_, conn any) bool {
		closeErr := s.CloseClient(conn.(*Connection).ID())
		if closeErr != nil {
			s.SendError(closeErr)
		}
		return true
	})

	s.CloseErrors()
	s.SetIsRunning(false)

	if errors.Is(err, syscall.EINVAL) {
		return nil
	} else {
		return err
	}
}

func (s *RfcommServer) configureListener(bindPort uint16, clientLimit int) error {
	s.listener = 0

	sock, err := unix.Socket(syscall.AF_BLUETOOTH, syscall.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		return err
	}

	sockInfo := &unix.SockaddrRFCOMM{
		Addr:    [6]uint8{},
		Channel: uint8(bindPort),
	}

	err = unix.Bind(sock, sockInfo)
	if err != nil {
		_ = unix.Close(sock)
		return err
	}

	if clientLimit < 0 {
		clientLimit = 4096
	}
	err = unix.Listen(sock, clientLimit)
	if err != nil {
		_ = unix.Close(sock)
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.listenContext = ctx
	s.listenCancelFunc = cancel

	s.listener = sock

	return nil
}

func (s *RfcommServer) listenForClientConnections() {
	for {
		conn, addr, acceptErr := unix.Accept(s.listener)
		if acceptErr != nil {
			if errors.Is(acceptErr, net.ErrClosed) {
				return
			}

			s.SendError(acceptErr)
			continue
		}

		go func() {
			addClientErr := s.addClientConnection(conn, addr)
			if addClientErr != nil {
				s.SendError(addClientErr)
			}
		}()
	}
}

func (s *RfcommServer) addClientConnection(rfcommConn int, address unix.Sockaddr) error {
	cfg := s.Config()

	clientCount := s.ClientCount()
	if cfg.ClientConnectionLimit > -1 && clientCount >= cfg.ClientConnectionLimit {
		return fmt.Errorf("%w : %d", comerr.ErrConnectionLimitReached, clientCount)
	}

	err := s.setConnectionOptions(rfcommConn, cfg.ReadTimeoutUs)
	if err != nil {
		return err
	}

	remoteAddress, err := rfcomm.ByteArrayToMacString(address.(*unix.SockaddrRFCOMM).Addr)
	if err != nil {
		return err
	}

	conn := &Connection{}
	conn.ConfigureRFCOMM(s, rfcommConn, remoteAddress, int64(cfg.IdleConnectionTimeoutMs), s.CloseClient)

	s.connections.Store(conn.ID(), conn)

	select {
	case s.newConnChan <- conn:
		break
	default:
	}

	return nil
}

func (s *RfcommServer) setConnectionOptions(connection int, readTimeoutUs int) error {
	if connection < 0 {
		return net.ErrClosed
	}

	// The microsecond part cannot exceed domain limit
	usRemainder := readTimeoutUs % 1000000
	sec := readTimeoutUs / 1000000

	tv := unix.Timeval{Sec: int64(sec), Usec: int64(usRemainder)}
	err := unix.SetsockoptTimeval(connection, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &tv)
	if err != nil {
		return fmt.Errorf("%w : %v", comerr.ErrSetReadTimeout, err)
	}

	err = unix.SetNonblock(connection, true)
	if err != nil {
		return fmt.Errorf("%w : %v", comerr.ErrSetNonBlockingMode, err)
	}

	closeLingerTime := unix.Linger{Onoff: 1, Linger: 0}
	err = unix.SetsockoptLinger(connection, unix.SOL_SOCKET, unix.SO_LINGER, &closeLingerTime)
	if err != nil {
		return fmt.Errorf("%w : %v", comerr.ErrSetLingerTimeout, err)
	}

	return nil
}

func (s *RfcommServer) read(conn *Connection, buffer []byte) (int, error) {
	if conn == nil || conn.rfcommConn == 0 {
		return -1, net.ErrClosed
	}

	count, err := unix.Read(conn.rfcommConn, buffer)

	err = socket.SinkReadWriteError(err)
	if err != nil {
		closeErr := s.CloseClient(conn.ID())
		if closeErr != nil {
			s.SendError(closeErr)
		}
	}

	return count, err
}

func (s *RfcommServer) write(conn *Connection, data []byte) (int, error) {
	if conn == nil || conn.rfcommConn < 1 {
		return -1, net.ErrClosed
	}

	count, err := unix.Write(conn.rfcommConn, data)

	err = socket.SinkReadWriteError(err)
	if err != nil {
		closeErr := s.CloseClient(conn.ID())
		if closeErr != nil {
			s.SendError(closeErr)
		}
	}

	return count, err
}

func (s *RfcommServer) handleListenCancel() {
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
