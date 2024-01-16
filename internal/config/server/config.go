package server

const (
	defaultClientConnectionLimit   = -1      // <0 means 4096, 0 means none
	defaultIdleConnectionTimeoutMs = 60000   // <1 means no idle connection pruning
	defaultErrorChanBufferSize     = 100     // error count
	defaultReadBufferSize          = 1500    // byte count, should match transport MTU
	defaultReadTimeoutUs           = 1000000 // <600 is essentially non-blocking
	defaultConnectionless          = false   // if true uses UDP instead of TCP
)

type Config struct {
	Address                 string
	Port                    uint16
	ClientConnectionLimit   int
	IdleConnectionTimeoutMs int
	ErrorChanBufferSize     int
	ReadBufferSize          int
	ReadTimeoutUs           int
	Connectionless          bool
}

func NewConfig(address string, port uint16) Config {
	cfg := Config{
		Address:                 address,
		Port:                    port,
		ClientConnectionLimit:   defaultClientConnectionLimit,
		IdleConnectionTimeoutMs: defaultIdleConnectionTimeoutMs,
		ErrorChanBufferSize:     defaultErrorChanBufferSize,
		ReadBufferSize:          defaultReadBufferSize,
		ReadTimeoutUs:           defaultReadTimeoutUs,
		Connectionless:          defaultConnectionless,
	}
	return cfg
}
