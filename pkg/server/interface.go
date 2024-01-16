package server

import (
	"io"
	"tonysoft.com/comm/internal/comerr"
	"tonysoft.com/comm/internal/comobj"
	"tonysoft.com/comm/internal/config"
	_config "tonysoft.com/comm/internal/config/server"
	_server "tonysoft.com/comm/internal/server"
	"tonysoft.com/comm/internal/socket"
	"tonysoft.com/comm/internal/transport"
)

// Server Public interface for working with instances of Server
// Thread-safe ✓
type Server interface {
	config.Configurable[_config.Config]
	Start() error
	Stop()
	comobj.Runnable
	ClientCount() int
	CloseClient(socket.ConnectionID) error
	Accept() <-chan socket.Connection
	comerr.Producer
}

// New Create a new instance of Server
func New(cfg _config.Config) (Server, error) {
	transportType, err := transport.GetTypeFromAddress(cfg.Address)
	if err != nil {
		return nil, err
	}

	var s Server

	switch transportType {
	case transport.TCP:
		if cfg.Connectionless {
			s = &_server.UdpServer{}
		} else {
			s = &_server.TcpServer{}
		}
	case transport.RFCOMM:
		s = &_server.RfcommServer{}
	}

	s.SetConfig(cfg)

	return s, nil
}

type Connection interface {
	ID() socket.ConnectionID
	RemoteAddress() string
	comobj.Connectable
	comobj.Idleable
	io.Reader
	io.Writer
	io.Closer
}
