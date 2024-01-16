/*******************************************************************************
 Message packet structure sent over the network:
 |<----------------------HEADER------------------------>|<----------PAYLOAD----------->|
 | 0 : 5 | 6 : 7 | 8  | 9 : 12 | 13 : 25 | 26 : 29 | 30 | 31 : {30+PAYSZ} | {31+PAYSZ} |
 | SYNC  | REPLY | MS |   ID   |SENT/RECV|  PAYSZ  | HC |     PAYLOAD     |     PC     |


 Label     | Size | Description
 -------------------------------------------------------------------------------
 SYNC        6      sync byte (22), repeated 6 times to form packet preamble
 REPLY       2      reply port, uint16 (port the node listens to for messages*)
 MS          1      message status, (see statuses below)
 ID          4      message ID, uint32
 SENT/RECV   13     sent/received timestamp, []byte (ASCII**), unix milliseconds
 PAYSZ       4      payload size, uint32
 HC          1      header checksum, { (∑REPLY + MS + ∑ID + ∑SENT + ∑PAYSZ) ^ 255 }
 PAYLOAD     n      payload, []byte (base-64)
 PC          1      payload checksum, { ∑PAYLOAD ^ 255 }

 * Message receipts sent by the callee/recipient are sent over the same
   connection on which the received message was sent by the caller/sender (i.e.,
   receipts are sent on the connection established by the caller).  That means
   the actual reply/destination port used by the callee to send receipts will be
   the caller's source port (chosen by the NIC) used for the underlying TCP
   connection. That is a low-level concern we don't have to worry about.  For
   regular messages (not receipts), the sender must always use the recipient's
   user-assigned address port as the destination port because that's the one the
   recipient is listening to for new messages.  Since a callee may want to
   initiate a new message in response to a message just received (i.e., the
   callee wants to reply to the caller beyond the automatically sent receipt
   for the first message received), the callee must know the socket the caller
   is bound to for receiving new messages.  Because the TCP connection already
   provides the IP address for the caller (as well as the source port, which
   is not needed in this case), all the callee needs to send a reply to the
   caller is the caller's "reply" port... the port number that the caller is
   bound to for new incoming messages and replies.  This is why the reply port
   is included with every message sent between nodes.

   New messages, which may have large payloads, are sent over a connection that
   the sender (caller node) "owns" and is generally responsible for closing once
   done sending messages.  Nodes, including callees, may terminate idle
   connections as per their configuration, but otherwise it's the expectation
   that callers can and will terminate the connection whenever they wish,
   that is why new messages sent by a calling node are not sent over a connection
   established by another node but rather one that the caller owns.  Receipts
   are an exception, as if the original calling node terminates the connection
   immediately after sending a message but before the callee sends the receipt
   then it's assumed that the caller does not care to receive such a receipt.
   Callers that expect receipts wait for a configurable period of time to
   receive them after sending messages and the callees/recipients will
   automatically send them unless configured not to do so.

   If the message is a regular/new message, then SENT/RECV will be equal to
   the time (Unix epoch, UTC, milliseconds) that the message was sent by the
   caller.  If the message is a receipt then SENT/RECV will be the time that
   the callee received the message.  The delta can be used to calculate
   network latency (if sending small messages, preferably with no payload)
   and bandwidth (if sending a message with a large payload, as the received
   timestamp is only made after the entire message has been received).

 ** SENT/RECV is sent as text to avoid chance it, combined with adjacent
    fields, would form the byte sequence used as the packet preamble.
    Likewise, the PAYLOAD is base-64 encoded, thus byte value 22 (SYNC) will
    never be part of the PAYLOAD bytes that are sent over the network.  Also
    note that PAYLOAD/PC are optional (a packet can contain just a header,
    which is the case for message receipts).


 Message Statuses:
   - 100 message sent
   - 101 message receipt not received (this message type not sent over network)
   - 200 message received successfully (header/payload came through ok)
   - 201 payload not received successfully (just the header came through ok)

   Note 100-level messages are created by the sender/caller, whereas 200-level
   messages (receipts) are created by the recipient/callee.
*******************************************************************************/

package node

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync/atomic"
	"time"
	"tonysoft.com/comm/pkg/comerr"
)

type MessageStatus byte

const (
	MessageSent        MessageStatus = 100
	MessageReceived                  = 200
	PayloadNotReceived               = 201
)

const (
	messageHeaderSize    = 31
	messageSyncByte      = 22
	messageSyncByteCount = 6
)

const (
	defaultChanBufferSize = 1024
)

var (
	messageNextId atomic.Uint32
)

/*******************************************************************************
 MESSAGE
*******************************************************************************/

type Message[T any] struct {
	id         uint32
	status     atomic.Uint32
	replyPort  uint16
	fromNode   string
	toNode     string
	sentOn     time.Time
	receivedOn time.Time

	Data *T
}

func (m *Message[T]) ID() uint32 {
	return m.id
}

func (m *Message[T]) Status() MessageStatus {
	return MessageStatus(m.status.Load())
}

func (m *Message[T]) FromNode() string {
	return m.fromNode
}

func (m *Message[T]) ToNode() string {
	return m.toNode
}

func (m *Message[T]) SentOn() time.Time {
	return m.sentOn
}

func (m *Message[T]) ReceivedOn() time.Time {
	return m.receivedOn
}

func (m *Message[T]) ToBytes() ([]byte, error) {
	// Calculate payload size and convert to bytes
	payloadSize := uint32(0)
	payloadChecksumSize := uint32(0)
	var payloadBytes []byte
	var payloadBytesBase64 []byte
	if m.Data != nil {
		if dataBytes, ok := any(m.Data).(*[]byte); ok {
			payloadBytes = *dataBytes
		} else {
			jsonBytes, err := json.Marshal(*m.Data)
			if err != nil {
				return nil, err
			}
			payloadBytes = jsonBytes
		}

		payloadBytesBase64 = make([]byte, base64.StdEncoding.EncodedLen(len(payloadBytes)))
		base64.StdEncoding.Encode(payloadBytesBase64, payloadBytes)
		payloadSize = uint32(len(payloadBytesBase64))
		payloadChecksumSize = 1
	}

	// Allocate memory for the entire message
	bytes := make([]byte, messageHeaderSize+payloadSize+payloadChecksumSize)

	// Set preamble
	for i := 0; i < messageSyncByteCount; i++ {
		bytes[i] = messageSyncByte
	}

	// Set reply port (the callee node will already know the source IP address and
	// source port, however only message receipts are sent to the source port, while
	// message replies are sent to the reply port).
	binary.BigEndian.PutUint16(bytes[6:8], m.replyPort)

	// Set message status
	bytes[8] = byte(m.status.Load())
	binary.BigEndian.PutUint32(bytes[9:13], m.id)

	// Set sent/received timestamp (received is used only for receipts)
	switch MessageStatus(m.status.Load()) {
	case MessageSent:
		sentOn := []byte(fmt.Sprintf("%013d", m.sentOn.UnixMilli()))
		for i := 13; i < 26; i++ {
			bytes[i] = sentOn[i-13]
		}
	case MessageReceived, PayloadNotReceived:
		receivedOn := []byte(fmt.Sprintf("%013d", m.receivedOn.UnixMilli()))
		for i := 13; i < 26; i++ {
			bytes[i] = receivedOn[i-13]
		}
	}

	// Set payload size
	binary.BigEndian.PutUint32(bytes[26:30], payloadSize)

	// Set header checksum
	bytes[30] = getChecksum(bytes[6:30])

	if payloadSize > 0 {
		// Set payload
		checksumIndex := int(messageHeaderSize + payloadSize)
		for i := messageHeaderSize; i < checksumIndex; i++ {
			bytes[i] = payloadBytesBase64[i-messageHeaderSize]
		}

		// Set payload checksum
		bytes[messageHeaderSize+payloadSize] = getChecksum(payloadBytesBase64)
	}

	return bytes, nil
}

func MessageFromBytes[T any](bytes []byte) (*Message[T], error) {
	bytesSize := len(bytes)

	// Messages must be at least as big as the (fixed-size) header
	if bytesSize < messageHeaderSize {
		return nil, comerr.ErrInvalidMessageFormat
	}

	// Validate the preamble
	for i := 0; i < messageSyncByteCount; i++ {
		if bytes[i] != messageSyncByte {
			return nil, comerr.ErrInvalidMessageFormat
		}
	}

	// Validate the header checksum
	if bytes[30] != getChecksum(bytes[6:30]) {
		return nil, comerr.ErrInvalidMessageFormat
	}

	msg := &Message[T]{}
	msg.replyPort = binary.BigEndian.Uint16(bytes[6:8])
	msg.status.Store(uint32(bytes[8]))
	msg.id = binary.BigEndian.Uint32(bytes[9:13])

	timestamp, err := strconv.ParseInt(string(bytes[13:26]), 10, 64)
	if err != nil {
		return nil, err
	}

	switch MessageStatus(msg.status.Load()) {
	case MessageSent:
		msg.sentOn = time.UnixMilli(timestamp)
	case MessageReceived | PayloadNotReceived:
		msg.receivedOn = time.UnixMilli(timestamp)
	}

	payloadSize := binary.BigEndian.Uint32(bytes[26:30])
	if payloadSize > 0 {
		if uint32(bytesSize) != messageHeaderSize+payloadSize+1 {
			return msg, comerr.ErrInvalidMessagePayload
		}

		dataBytesBase64 := bytes[messageHeaderSize : bytesSize-1]

		if bytes[messageHeaderSize+payloadSize] != getChecksum(dataBytesBase64) {
			return msg, comerr.ErrInvalidMessagePayload
		}

		dataBytes := make([]byte, len(dataBytesBase64))
		decodedCount, b64Err := base64.StdEncoding.Decode(dataBytes, dataBytesBase64)
		if b64Err != nil {
			return msg, comerr.ErrInvalidMessagePayload
		}
		dataBytes = dataBytes[:decodedCount]

		if _, ok := any(msg.Data).(*[]byte); ok {
			dataBytesT := any(dataBytes).(T)
			msg.Data = &dataBytesT
		} else {
			var data T
			jsonErr := json.Unmarshal(dataBytes, &data)
			if jsonErr != nil {
				return msg, comerr.ErrInvalidMessagePayload
			}
			msg.Data = &data
		}
	}

	return msg, nil
}

func NewMessage[T any](replyPort uint16, toNode string, data *T) *Message[T] {
	msg := &Message[T]{
		id:        messageNextId.Add(1),
		replyPort: replyPort,
		toNode:    toNode,
		sentOn:    time.UnixMilli(time.Now().UTC().UnixMilli()), // trim nanoseconds
		Data:      data,
	}
	msg.status.Store(uint32(MessageSent))
	return msg
}

func NewMessageReceipt[T any](id uint32, replyPort uint16, toNode string, status MessageStatus) *Message[T] {
	rcpt := &Message[T]{
		id:         id,
		replyPort:  replyPort,
		toNode:     toNode,
		receivedOn: time.UnixMilli(time.Now().UTC().UnixMilli()), // trim nanoseconds
	}
	rcpt.status.Store(uint32(status))
	return rcpt
}

func getChecksum(byteSlice []byte) byte {
	checksum := byte(0)
	for _, b := range byteSlice {
		checksum += b
	}
	return checksum ^ 255
}

/*******************************************************************************
 STREAM
*******************************************************************************/

type MessageStream[T any] struct {
	reader      io.Reader
	streamChan  chan *Message[T]
	processChan chan bool
	err         atomic.Value
}

func (s *MessageStream[T]) Stream(reader io.Reader, bufferSize ...int) <-chan *Message[T] {
	if s.reader == nil {
		s.reader = reader
		if len(bufferSize) > 0 {
			s.streamChan = make(chan *Message[T], bufferSize[0])
		} else {
			s.streamChan = make(chan *Message[T], defaultChanBufferSize)
		}

		go s.process()
	}

	return s.streamChan
}

func (s *MessageStream[T]) Error() error {
	e := s.err.Load()
	if e != nil {
		return e.(error)
	}
	return nil
}

func (s *MessageStream[T]) Close() error {
	if s.processChan != nil {
		select {
		case <-s.processChan:
			break
		default:
			close(s.processChan)
		}
	}
	return s.Error()
}

func (s *MessageStream[T]) process() {
	s.processChan = make(chan bool)
	buffer := make([]byte, 1500)
	mb := NewMessageBuilder[T]()

	for {
		select {
		case <-s.processChan:
			s.processChan = nil
			close(s.streamChan)
			return
		default:
		}

		count, err := s.reader.Read(buffer)
		if err != nil {
			s.err.Store(err)
			close(s.streamChan)
			return
		}

		for i := 0; i < count; i++ {
			b := buffer[i]
			if mb.WriteByte(b) {
				msg, e := mb.Message()
				if e == nil {
					msg.status.Store(MessageReceived)

					select {
					case s.streamChan <- msg:
						break
					default:
					}
				} else {
					if e == comerr.ErrInvalidMessagePayload {
						msg.status.Store(PayloadNotReceived)
					}

					s.err.Store(e)
					close(s.streamChan)
					return
				}
			}
		}
	}
}

func NewMessageStream[T any](reader io.Reader, bufferSize ...int) <-chan *Message[T] {
	ms := &MessageStream[T]{}
	return ms.Stream(reader, bufferSize...)
}

/*******************************************************************************
 BUILDER
*******************************************************************************/

type MessageBuilder[T any] struct {
	preamble          []byte
	preambleEndIndex  int
	payloadStartIndex int
	payloadEndIndex   int
	pointer           int

	inPreamble bool
	inHeader   bool
	inPayload  bool

	Data     []byte
	DataSize int
}

// WriteByte Add a byte to the internal buffer, returning true if the
// builder is ready to produce a complete node.Message via Message().
//
// Note the name WriteByte was named after strings.Builder's version to
// provide a familiar API, however they do not share the same signature,
// which generates the GoStandardMethods warning disabled below.
//
//goland:noinspection GoStandardMethods
func (m *MessageBuilder[T]) WriteByte(b byte) bool {
	if m.inPayload {
		if m.pointer == m.payloadEndIndex {
			m.Data[m.pointer] = b
			m.DataSize = m.pointer + 1
			m.Reset()
			return true
		} else {
			m.Data[m.pointer] = b
			m.pointer++
		}
	} else if m.inHeader {
		m.Data[m.pointer] = b
		m.pointer++
		if m.pointer == m.payloadStartIndex {
			cs := byte(0)
			for i := 6; i < 30; i++ {
				cs += m.Data[i]
			}
			cs ^= 255

			if cs != m.Data[30] {
				m.Reset()
			} else {
				payloadSize := binary.BigEndian.Uint32(m.Data[26:30])
				if payloadSize > 0 {
					m.payloadEndIndex = 31 + int(payloadSize)
					payloadBuffer := make([]byte, payloadSize+1)
					m.Data = append(m.Data, payloadBuffer...)
					m.inPayload = true
					m.inHeader = false
				} else {
					m.DataSize = m.pointer
					m.Reset()
					return true
				}
			}
		}
	} else if m.inPreamble {
		if m.preamble[m.pointer] == b {
			m.Data[m.pointer] = b
			m.pointer++

			if m.pointer > m.preambleEndIndex {
				m.inHeader = true
				m.inPreamble = false
			}
		} else {
			m.Reset()
		}
	} else {
		if m.preamble[m.pointer] == b {
			m.Data[m.pointer] = b
			m.pointer++
			m.inPreamble = true
		}
	}

	return false
}

// Reset Resets the internal state machine logic, which is done automatically
// when WriteByte() returns true or when an unexpected byte is sent via WriteByte().
func (m *MessageBuilder[T]) Reset() {
	m.pointer = 0
	m.inPreamble = false
	m.inHeader = false
	m.inPayload = false
}

// Message This can be called once WriteByte() returns true, indicating a complete
// message has been received via WriteByte().
func (m *MessageBuilder[T]) Message() (*Message[T], error) {
	return MessageFromBytes[T](m.Data[:m.DataSize])
}

func NewMessageBuilder[T any]() *MessageBuilder[T] {
	return &MessageBuilder[T]{
		preamble:          []byte{22, 22, 22, 22, 22, 22},
		preambleEndIndex:  5,
		payloadStartIndex: 31,
		Data:              make([]byte, 31),
	}
}
