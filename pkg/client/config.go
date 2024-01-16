package client

import _config "tonysoft.com/comm/internal/config/client"

func NewConfig(remoteAddress string, remotePort uint16, connectionless ...bool) _config.Config {
	useUdp := false
	if len(connectionless) > 0 {
		useUdp = connectionless[0]
	}
	cfg := _config.NewConfig(remoteAddress, remotePort)
	cfg.Connectionless = useUdp
	return cfg
}
