package server

import (
	"context"
	"sync"
	"tonysoft.com/comm/internal/comerr"
	"tonysoft.com/comm/internal/comobj"
	"tonysoft.com/comm/internal/config"
	_server "tonysoft.com/comm/internal/config/server"
	"tonysoft.com/comm/internal/socket"
)

type BaseServer struct {
	config.DefaultConfigurable[_server.Config]
	comerr.DefaultProducer
	comobj.DefaultRunnable

	listenContext    context.Context
	listenCancelFunc context.CancelFunc

	connections sync.Map // map[socket.ConnectionID]*Connection
	newConnChan chan socket.Connection
}

func (s *BaseServer) Stop() {
	if !s.IsRunning() {
		return
	}

	if s.listenCancelFunc != nil {
		s.listenCancelFunc()
	}
}

func (s *BaseServer) Accept() <-chan socket.Connection {
	return s.newConnChan
}

func (s *BaseServer) ClientCount() int {
	count := 0
	s.connections.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

func (s *BaseServer) clearConnections() {
	s.connections.Range(func(key, _ any) bool {
		s.connections.Delete(key)
		return true
	})

	clientLimit := s.Config().ClientConnectionLimit
	if clientLimit < 0 {
		clientLimit = 4096
	}
	s.newConnChan = make(chan socket.Connection, clientLimit)
}
