package node

import (
	"sync"
	"tonysoft.com/comm/internal/comerr"
	"tonysoft.com/comm/internal/comobj"
	"tonysoft.com/comm/internal/config"
	_node "tonysoft.com/comm/internal/config/node"
	"tonysoft.com/comm/pkg/server"
)

type BaseNode[T any] struct {
	config.DefaultConfigurable[_node.Config]
	comobj.DefaultRunnable

	server       server.Server
	replyAddress string
	replyPort    uint16

	connections sync.Map // map[socket.ConnectionID]*Connection

	incomingChan chan *Message[T]
	statusChan   chan *Message[T]

	comerr.DefaultProducer
}
