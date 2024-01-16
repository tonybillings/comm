package client

import (
	"io"
	_client "tonysoft.com/comm/internal/client"
	"tonysoft.com/comm/internal/comobj"
	"tonysoft.com/comm/internal/config"
	_config "tonysoft.com/comm/internal/config/client"
	"tonysoft.com/comm/internal/transport"
)

// Client Public interface for working with instances of Client
// Thread-safe ✓
type Client interface {
	config.Configurable[_config.Config]
	RemoteAddress() string
	Start() error
	Stop() error
	comobj.Connectable
	io.Reader
	io.Writer
}

// New Create a new instance of Client
func New(cfg _config.Config) (Client, error) {
	transportType, err := transport.GetTypeFromAddress(cfg.RemoteAddress)
	if err != nil {
		return nil, err
	}

	var c Client

	switch transportType {
	case transport.TCP:
		if cfg.Connectionless {
			c = &_client.UdpClient{}
		} else {
			c = &_client.TcpClient{}
		}
	case transport.RFCOMM:
		c = &_client.RfcommClient{}
	}

	c.SetConfig(cfg)

	return c, nil
}
