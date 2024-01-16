package transport

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"tonysoft.com/comm/pkg/comerr"
)

type Type int64

const (
	NotSet Type = iota
	TCP
	UDP
	RFCOMM
)

func GetTypeFromAddress(address string) (Type, error) {
	address = strings.TrimSpace(address)

	if address == "" {
		return NotSet, comerr.ErrAddressEmpty
	}

	if address == "localhost" || strings.HasPrefix(address, "localhost:") {
		return TCP, nil
	}

	_, err := netip.ParseAddr(address)
	if err == nil {
		return TCP, nil
	}

	if strings.HasPrefix(address, ":") {
		address = "0.0.0.0" + address
	}

	_, err = netip.ParseAddrPort(address)
	if err == nil {
		return TCP, nil
	}

	_, err = net.ParseMAC(address)
	if err == nil {
		return RFCOMM, nil
	}

	return NotSet, comerr.ErrAddressFormatUnknown
}

func GetHostAndPortFromTcpAddress(address string) (host string, port uint16, err error) {
	if addrType, e := GetTypeFromAddress(address); e != nil || addrType != TCP {
		return "", 0, comerr.ErrAddressFormatUnknown
	}

	addrParts := strings.Split(address, ":")
	if len(addrParts) != 2 {
		return "", 0, comerr.ErrAddressFormatUnknown
	}

	p, err := strconv.ParseUint(addrParts[1], 10, 16)
	if err != nil {
		return "", 0, comerr.ErrAddressFormatUnknown
	}

	port = uint16(p)

	host = addrParts[0]
	if host == "" {
		host = "0.0.0.0"
	}

	return host, port, nil
}

func GetTcpAddressFromHostAndPort(host string, port uint16) (string, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	_, err := GetTypeFromAddress(address)
	return address, err
}
