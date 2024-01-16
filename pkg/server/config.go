package server

import _config "tonysoft.com/comm/internal/config/server"

func NewConfig(address string, port uint16, connectionless ...bool) _config.Config {
	useUdp := false
	if len(connectionless) > 0 {
		useUdp = connectionless[0]
	}
	cfg := _config.NewConfig(address, port)
	cfg.Connectionless = useUdp
	return cfg
}
