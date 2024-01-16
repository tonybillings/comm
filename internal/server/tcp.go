package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"syscall"
	"time"
	"tonysoft.com/comm/internal/socket"
	"tonysoft.com/comm/internal/transport"
	"tonysoft.com/comm/pkg/comerr"
)

type TcpServer struct {
	BaseServer

	listener      net.Listener
	readTimeoutUs int
}

func (s *TcpServer) Start() error {
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

	go s.listenForClientConnections()
	go s.handleListenCancel()

	s.SetIsRunning(true)

	return nil
}

func (s *TcpServer) CloseClient(id socket.ConnectionID) error {
	conn, ok := s.connections.Load(id)
	if !ok {
		return nil
	}

	s.connections.Delete(id)
	return conn.(*Connection).tcpConn.Close()
}

func (s *TcpServer) close() error {
	err := s.listener.Close()
	s.listener = nil
	s.listenContext = nil
	s.listenCancelFunc = nil

	s.connections.Range(func(_, conn any) bool {
		closeErr := s.CloseClient(conn.(*Connection).ID())
		if closeErr != nil {
			s.SendError(closeErr)
		}
		return true
	})

	close(s.newConnChan)
	s.CloseErrors()
	s.SetIsRunning(false)

	if errors.Is(err, syscall.EINVAL) {
		return nil
	} else {
		return err
	}
}

func (s *TcpServer) configureListener(bindAddress string) error {
	s.listener = nil

	addr, err := net.ResolveTCPAddr("tcp", bindAddress)
	if err != nil {
		return err
	}

	var cfg = net.ListenConfig{
		Control: socket.ControlFunc,
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.listenContext = ctx
	s.listenCancelFunc = cancel

	listener, err := cfg.Listen(s.listenContext, "tcp4", addr.String())
	if err != nil {
		return err
	}
	s.listener = listener

	return nil
}

func (s *TcpServer) listenForClientConnections() {
	for {
		conn, acceptErr := s.listener.Accept()
		if acceptErr != nil {
			if errors.Is(acceptErr, net.ErrClosed) {
				return
			}

			s.SendError(acceptErr)
			continue
		}

		if tcpConn, ok := conn.(*net.TCPConn); !ok {
			s.SendError(comerr.ErrConnectAborted)
			continue
		} else {
			go func() {
				addClientErr := s.addClientConnection(tcpConn)
				if addClientErr != nil {
					s.SendError(addClientErr)
				}
			}()
		}
	}
}

func (s *TcpServer) addClientConnection(netConn *net.TCPConn) error {
	cfg := s.Config()

	err := s.verifyConnectionLimit(cfg.ClientConnectionLimit)
	if err != nil {
		return err
	}

	err = s.setConnectionOptions(netConn)
	if err != nil {
		return err
	}

	conn := &Connection{}
	conn.ConfigureTCP(s, netConn, int64(cfg.IdleConnectionTimeoutMs), s.CloseClient)

	s.connections.Store(conn.ID(), conn)

	select {
	case s.newConnChan <- conn:
		break
	default:
	}

	return nil
}

func (s *TcpServer) verifyConnectionLimit(connectionLimit int) error {
	clientCount := s.ClientCount()
	if connectionLimit < 0 {
		connectionLimit = 4096
	}
	if clientCount >= connectionLimit {
		return fmt.Errorf("%w : %d", comerr.ErrConnectionLimitReached, clientCount)
	}

	return nil
}

func (s *TcpServer) read(conn *Connection, buffer []byte) (int, error) {
	if conn == nil || conn.tcpConn == nil {
		return -1, net.ErrClosed
	}

	err := conn.tcpConn.SetReadDeadline(time.Now().Add(time.Duration(s.readTimeoutUs) * time.Microsecond))
	if err != nil {
		return -1, fmt.Errorf("%w : %v", comerr.ErrSetReadTimeout, err)
	}

	count, err := conn.tcpConn.Read(buffer)

	err = socket.SinkReadWriteError(err)
	if err != nil {
		closeErr := s.CloseClient(conn.ID())
		if closeErr != nil {
			s.SendError(closeErr)
		}
	}

	return count, err
}

func (s *TcpServer) write(conn *Connection, data []byte) (int, error) {
	if conn == nil || conn.tcpConn == nil {
		return -1, net.ErrClosed
	}

	count, err := conn.tcpConn.Write(data)

	err = socket.SinkReadWriteError(err)
	if err != nil {
		closeErr := s.CloseClient(conn.ID())
		if closeErr != nil {
			s.SendError(closeErr)
		}
	}

	return count, err
}

func (s *TcpServer) setConnectionOptions(conn *net.TCPConn) error {
	err := conn.SetLinger(0)
	if err != nil {
		return fmt.Errorf("%w : %v", comerr.ErrSetLingerTimeout, err)
	}
	return nil
}

func (s *TcpServer) handleListenCancel() {
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
