package client

const (
	defaultConnectTimeoutSec = 30      // how long to wait for the server to answer
	defaultReadTimeoutUs     = 1000000 // <600 is essentially non-blocking
	defaultConnectionless    = false   // if true uses UDP instead of TCP
)

type Config struct {
	RemoteAddress     string
	RemotePort        uint16
	ConnectTimeoutSec int
	ReadTimeoutUs     int
	Connectionless    bool
}

func NewConfig(remoteAddress string, remotePort uint16) Config {
	cfg := Config{
		RemoteAddress:     remoteAddress,
		RemotePort:        remotePort,
		ConnectTimeoutSec: defaultConnectTimeoutSec,
		ReadTimeoutUs:     defaultReadTimeoutUs,
		Connectionless:    defaultConnectionless,
	}
	return cfg
}
