package node

import (
	"fmt"
	"time"
	"tonysoft.com/comm/internal/socket"
	"tonysoft.com/comm/internal/transport"
	"tonysoft.com/comm/pkg/client"
	"tonysoft.com/comm/pkg/comerr"
	"tonysoft.com/comm/pkg/server"
)

type TcpNode[T any] struct {
	BaseNode[T]
}

func (n *TcpNode[T]) Start() error {
	if n.IsRunning() {
		return comerr.ErrNodeAlreadyRunning
	}

	cfg := n.Config()

	n.ConfigureErrors(cfg.ErrorChanBufferSize)
	n.incomingChan = make(chan *Message[T], cfg.RecvChanBufferSize)
	n.statusChan = make(chan *Message[T], cfg.StatusChanBufferSize)

	err := n.startServer(cfg.Address, cfg.IncomingConnectionLimit, cfg.IdleConnectionTimeoutMs, cfg.SendMessageReceipts)
	if err != nil {
		return err
	}

	n.SetIsRunning(true)

	return nil
}

func (n *TcpNode[T]) Stop() {
	if !n.IsRunning() {
		return
	}

	if n.server != nil {
		n.server.Stop()
	}

	n.connections.Range(func(id any, conn any) bool {
		err := conn.(*Connection).Close()
		if err != nil {
			n.SendError(err)
		}
		n.connections.Delete(id)
		return true
	})

	n.SetIsRunning(false)

	if n.incomingChan != nil {
		select {
		case <-n.incomingChan:
			return
		default:
			close(n.incomingChan)
		}
	}
}

func (n *TcpNode[T]) ConnectionCount() int {
	count := 0
	n.connections.Range(func(_ any, _ any) bool {
		count++
		return true
	})
	return count
}

func (n *TcpNode[T]) ConnectedNodes() []string {
	nodes := make([]string, 0)
	n.connections.Range(func(_ any, value any) bool {
		nodes = append(nodes, value.(*Connection).RemoteAddress())
		return true
	})
	return nodes
}

func (n *TcpNode[T]) Send(toNode string, data *T) (*Message[T], error) {
	msg := NewMessage[T](n.replyPort, toNode, data)

	conn := n.getConnectionByAddress(toNode)
	if conn == nil {
		c, err := n.addOutgoingConnection(toNode)
		if err != nil {
			return nil, err
		}
		conn = c
	}

	msgBytes, err := msg.ToBytes()
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(msgBytes)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (n *TcpNode[T]) Recv() <-chan *Message[T] {
	return n.incomingChan
}

func (n *TcpNode[T]) Status() <-chan *Message[T] {
	return n.statusChan
}

func (n *TcpNode[T]) startServer(address string, connectionLimit int, idleConnTimeoutMs int, sendReceipts bool) error {
	host, port, err := transport.GetHostAndPortFromTcpAddress(address)
	if err != nil {
		return err
	}

	n.replyAddress = fmt.Sprintf("%s:%d", host, port)
	n.replyPort = port

	serverCfg := server.NewConfig(host, port)
	serverCfg.ClientConnectionLimit = connectionLimit
	serverCfg.IdleConnectionTimeoutMs = idleConnTimeoutMs

	s, err := server.New(serverCfg)
	if err != nil {
		return err
	}
	n.server = s

	err = s.Start()
	if err != nil {
		return err
	}

	go func() {
		for conn := range s.Accept() {
			go n.handleIncomingConnection(conn, serverCfg.IdleConnectionTimeoutMs, sendReceipts)
		}
	}()

	go n.pruneIdleConnections()

	return nil
}

func (n *TcpNode[T]) handleIncomingConnection(conn socket.Connection, idleTimeoutMs int, sendReceipts bool) {
	callerHost, _, err := transport.GetHostAndPortFromTcpAddress(conn.RemoteAddress())
	if err != nil {
		n.SendError(err)
		return
	}

	c := NewConnection(conn.RemoteAddress(), int64(idleTimeoutMs), n.closeConnection)
	defer func() {
		e := c.Close()
		if e != nil {
			n.SendError(e)
		}
	}()

	c.connectionType = Callee
	c.calleeConn = conn

	n.connections.Store(c.ID(), c)

	// Receive incoming messages until the connection is closed
	for msg := range NewMessageStream[T](conn) {
		msg.receivedOn = time.Now().UTC()
		msg.fromNode = fmt.Sprintf("%s:%d", callerHost, msg.replyPort)
		msg.toNode = n.replyAddress

		n.incomingChan <- msg

		if !n.IsRunning() {
			close(n.incomingChan)
			return
		}

		if sendReceipts {
			e := n.sendReceipt(msg, conn)
			if e != nil {
				n.SendError(e)
			}
		}
	}
}

func (n *TcpNode[T]) sendReceipt(message *Message[T], conn socket.Connection) error {
	rcpt := NewMessageReceipt[T](message.ID(), n.replyPort, conn.RemoteAddress(), message.Status())

	rcptBytes, err := rcpt.ToBytes()
	if err != nil {
		return err
	}

	_, err = conn.Write(rcptBytes)
	if err != nil {
		return err
	}

	return nil
}

func (n *TcpNode[T]) addOutgoingConnection(toNode string) (*Connection, error) {
	cfg := n.Config()

	err := n.verifyConnectionLimit(cfg.OutgoingConnectionLimit)
	if err != nil {
		return nil, err
	}

	calleeHost, calleePort, err := transport.GetHostAndPortFromTcpAddress(toNode)
	if err != nil {
		return nil, err
	}

	clientCfg := client.NewConfig(calleeHost, calleePort)
	c, err := client.New(clientCfg)
	if err != nil {
		return nil, err
	}

	err = c.Start()
	if err != nil {
		return nil, err
	}

	conn := NewConnection(toNode, int64(cfg.IdleConnectionTimeoutMs), n.closeConnection)
	conn.connectionType = Caller
	conn.callerConn = c

	n.connections.Store(conn.ID(), conn)

	go func() {
		// Receive incoming message receipts until the connection is closed
		for rcpt := range NewMessageStream[T](c) {
			rcpt.receivedOn = time.Now().UTC()
			rcpt.fromNode = fmt.Sprintf("%s:%d", calleeHost, rcpt.replyPort)
			rcpt.toNode = n.replyAddress

			select {
			case n.statusChan <- rcpt:
				break
			default:
			}

			if !n.IsRunning() {
				close(n.statusChan)
				return
			}
		}

		select {
		case <-n.statusChan:
			break
		default:
			close(n.statusChan)
		}
	}()

	return conn, nil
}

func (n *TcpNode[T]) verifyConnectionLimit(connectionLimit int) error {
	if connectionLimit < 0 {
		connectionLimit = 4096
	}
	if n.server.ClientCount() >= connectionLimit {
		return comerr.ErrConnectionLimitReached
	}
	return nil
}

func (n *TcpNode[T]) getConnectionByAddress(address string) *Connection {
	var conn *Connection
	n.connections.Range(func(_ any, value any) bool {
		c := value.(*Connection)
		if c.RemoteAddress() == address {
			conn = c
			return false
		}
		return true
	})
	return conn
}

func (n *TcpNode[T]) closeConnection(id socket.ConnectionID) error {
	if conn, ok := n.connections.Load(id); ok {
		c := conn.(*Connection)

		n.connections.Delete(id)

		if c.calleeConn != nil {
			return c.calleeConn.Close()
		} else if c.callerConn != nil {
			return c.callerConn.Stop()
		}
	}
	return nil
}

func (n *TcpNode[T]) pruneIdleConnections() {
	for {
		n.connections.Range(func(key any, value any) bool {
			conn := value.(*Connection)
			if conn.IsIdle() {
				err := conn.Close()
				if err != nil {
					n.SendError(err)
				}
			}
			return true
		})

		time.Sleep(500 * time.Millisecond)

		if n == nil || !n.IsRunning() {
			return
		}
	}
}
