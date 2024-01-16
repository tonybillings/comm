package node

import _config "tonysoft.com/comm/internal/config/node"

func NewConfig(nodeAddress string) _config.Config {
	return _config.NewConfig(nodeAddress)
}
