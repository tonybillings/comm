package comerr

import "errors"

const (
	NotImplemented         = "function/feature not implemented"
	SetReadTimeout         = "failed to set read timeout"
	SetLingerTimeout       = "failed to set linger timeout"
	SetNonBlockingMode     = "failed to set socket in non-blocking mode"
	ConnectAborted         = "connection attempt aborted"
	ConnectTimeout         = "could not connect within timeout period"
	DisconnectTimeout      = "could not disconnect within timeout period"
	ParseMacAddress        = "could not parse MAC address"
	AddressEmpty           = "address is empty"
	AddressFormatUnknown   = "address does not match a known format"
	ClientAlreadyConnected = "client is already connected"
	ServerAlreadyRunning   = "server is already running"
	NodeAlreadyRunning     = "node is already running"
	ConnectionLimitReached = "connection limit reached"
	InvalidMessageFormat   = "message could not be instantiated from bytes"
	InvalidMessagePayload  = "message payload is missing or corrupt"
)

var (
	ErrNotImplemented         = errors.New(NotImplemented)
	ErrSetReadTimeout         = errors.New(SetReadTimeout)
	ErrSetLingerTimeout       = errors.New(SetLingerTimeout)
	ErrSetNonBlockingMode     = errors.New(SetNonBlockingMode)
	ErrConnectAborted         = errors.New(ConnectAborted)
	ErrConnectTimeout         = errors.New(ConnectTimeout)
	ErrDisconnectTimeout      = errors.New(DisconnectTimeout)
	ErrParseMacAddress        = errors.New(ParseMacAddress)
	ErrAddressEmpty           = errors.New(AddressEmpty)
	ErrAddressFormatUnknown   = errors.New(AddressFormatUnknown)
	ErrClientAlreadyConnected = errors.New(ClientAlreadyConnected)
	ErrServerAlreadyRunning   = errors.New(ServerAlreadyRunning)
	ErrNodeAlreadyRunning     = errors.New(NodeAlreadyRunning)
	ErrConnectionLimitReached = errors.New(ConnectionLimitReached)
	ErrInvalidMessageFormat   = errors.New(InvalidMessageFormat)
	ErrInvalidMessagePayload  = errors.New(InvalidMessagePayload)
)
