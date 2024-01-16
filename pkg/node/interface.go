package node

import (
	"tonysoft.com/comm/internal/comerr"
	"tonysoft.com/comm/internal/comobj"
	"tonysoft.com/comm/internal/config"
	_config "tonysoft.com/comm/internal/config/node"
	_node "tonysoft.com/comm/internal/node"
	"tonysoft.com/comm/internal/transport"
)

// Node Public interface for working with instances of Node[T]
// Thread-safe ✓
type Node[T any] interface {
	config.Configurable[_config.Config]
	Start() error
	Stop()
	comobj.Runnable
	ConnectionCount() int
	ConnectedNodes() []string
	Send(string, *T) (*_node.Message[T], error)
	Recv() <-chan *_node.Message[T]
	Status() <-chan *_node.Message[T]
	comerr.Producer
}

// New Create a new instance of Node[T]
func New[T any](cfg _config.Config) (Node[T], error) {
	transportType, err := transport.GetTypeFromAddress(cfg.Address)
	if err != nil {
		return nil, err
	}

	var n Node[T]

	switch transportType {
	case transport.TCP:
		n = &_node.TcpNode[T]{}
	case transport.RFCOMM:
		panic("the Node API does not support RFCOMM")
	}

	n.SetConfig(cfg)

	return n, nil
}

// Message Use this function to cast any to Message[T] or to
// get a nil pointer to Message[T].
func Message[T any](message any) *_node.Message[T] {
	if message == nil {
		var ptr *_node.Message[T]
		return ptr
	}
	return message.(*_node.Message[T])
}

// MessageStatus Export the internal enum used to track message status
type MessageStatus byte

const (
	MessageSent        = _node.MessageSent
	MessageReceived    = _node.MessageReceived
	PayloadNotReceived = _node.PayloadNotReceived
)
