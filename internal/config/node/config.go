package node

const (
	defaultIncomingConnectionLimit = -1      // <0 means 4096, 0 means none
	defaultOutgoingConnectionLimit = -1      // <0 means 4096, 0 means none
	defaultConnectTimeoutSec       = 30      // how long to wait for the callee to answer
	defaultIdleConnectionTimeoutMs = 60000   // <1 means no idle connection pruning
	defaultErrorChanBufferSize     = 100     // error count
	defaultRecvChanBufferSize      = 100     // *Message[T] count
	defaultStatusChanBufferSize    = 100     // *Message[T] count
	defaultReadBufferSize          = 1500    // byte count, should match transport MTU
	defaultReadTimeoutUs           = 1000000 // <600 is essentially non-blocking
	defaultSendMessageReceipts     = true    // Automatically send a receipt upon receiving a message
)

type Config struct {
	Address                 string
	IncomingConnectionLimit int
	OutgoingConnectionLimit int
	ConnectTimeoutSec       int
	IdleConnectionTimeoutMs int
	ErrorChanBufferSize     int
	RecvChanBufferSize      int
	StatusChanBufferSize    int
	ReadBufferSize          int
	ReadTimeoutUs           int
	SendMessageReceipts     bool
}

func NewConfig(address string) Config {
	cfg := Config{
		Address:                 address,
		IncomingConnectionLimit: defaultIncomingConnectionLimit,
		OutgoingConnectionLimit: defaultOutgoingConnectionLimit,
		ConnectTimeoutSec:       defaultConnectTimeoutSec,
		IdleConnectionTimeoutMs: defaultIdleConnectionTimeoutMs,
		ErrorChanBufferSize:     defaultErrorChanBufferSize,
		RecvChanBufferSize:      defaultRecvChanBufferSize,
		StatusChanBufferSize:    defaultStatusChanBufferSize,
		ReadBufferSize:          defaultReadBufferSize,
		ReadTimeoutUs:           defaultReadTimeoutUs,
		SendMessageReceipts:     defaultSendMessageReceipts,
	}
	return cfg
}
