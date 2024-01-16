package test

import (
	"testing"
	"tonysoft.com/comm/internal/transport"
)

func TestGetTransportTypeFromAddress(t *testing.T) {
	failTest := func(testNum int) {
		t.Errorf("unexpected result from GetTypeFromAddress(), test #%d", testNum)
	}

	address := "   "
	if tt, err := transport.GetTypeFromAddress(address); tt != transport.NotSet || err == nil {
		failTest(1)
	}

	address = "xxx"
	if tt, err := transport.GetTypeFromAddress(address); tt != transport.NotSet || err == nil {
		failTest(2)
	}

	address = "127.0.0.1"
	if tt, err := transport.GetTypeFromAddress(address); tt != transport.TCP || err != nil {
		failTest(3)
	}

	address = "localhost"
	if tt, err := transport.GetTypeFromAddress(address); tt != transport.TCP || err != nil {
		failTest(4)
	}

	address = "aa:11:bb:22:cc:33"
	if tt, err := transport.GetTypeFromAddress(address); tt != transport.RFCOMM || err != nil {
		failTest(5)
	}
}
