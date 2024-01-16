package client

import (
	"tonysoft.com/comm/internal/comobj"
	"tonysoft.com/comm/internal/config"
	_config "tonysoft.com/comm/internal/config/client"
)

type BaseClient struct {
	config.DefaultConfigurable[_config.Config]
	comobj.DefaultConnectable
}

func (c *BaseClient) RemoteAddress() string {
	return c.Config().RemoteAddress
}
