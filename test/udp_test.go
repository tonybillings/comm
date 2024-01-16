package test

import (
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"tonysoft.com/comm/pkg/client"
	"tonysoft.com/comm/pkg/server"
)

const (
	useConnectionless = true
)

func TestUdpServerReadWrite(t *testing.T) {
	testPassed := false
	testPing := []byte("ping")
	testPong := []byte("pong")

	serverCfg := server.NewConfig(net.IPv4zero.String(), 8375, useConnectionless)
	s, err := server.New(serverCfg)
	if err != nil {
		t.Error(err)
		return
	}

	err = s.Start()
	if err != nil {
		t.Error(err)
		return
	}
	defer s.Stop()

	go func() {
		for conn := range s.Accept() {
			request := make([]byte, 4)
			_, e := conn.Read(request)
			if e != nil {
				t.Errorf("connection read error: %v", e)
				return
			}

			if string(request) == string(testPing) {
				testPassed = true
			}

			_, err = conn.Write(testPong)
			if err != nil {
				t.Errorf("connection write error: %v", err)
				return
			}
		}
	}()

	clientCfg := client.NewConfig(net.IPv4zero.String(), 8375, useConnectionless)
	c, err := client.New(clientCfg)
	if err != nil {
		t.Error(err)
		return
	}

	err = c.Start()
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		e := c.Stop()
		if e != nil {
			t.Error(e)
		}
	}()

	request := testPing
	count, err := c.Write(request)
	if err != nil {
		t.Error(err)
		return
	}
	if count != len(request) {
		t.Error(fmt.Errorf("expected to write %d bytes, wrote %d instead", len(request), count))
		return
	}

	time.Sleep(time.Second)

	buffer := make([]byte, len(testPong))
	count, err = c.Read(buffer)
	if err != nil {
		t.Error(err)
		return
	}
	if count != len(testPong) {
		t.Error(fmt.Errorf("expected to read %d bytes, read %d instead", len(testPong), count))
		return
	}
	if string(buffer) != string(testPong) {
		t.Error(fmt.Errorf("expected to receive '%s', received '%s' instead", string(testPong), string(buffer)))
		return
	}

	if !testPassed {
		t.Error("ping/pong test failed")
		return
	}
}

func TestUdpServerMultiClient(t *testing.T) {
	const testCount = 5
	const clientCount = 50
	const requestLength = 500

	startingRoutineCount := runtime.NumGoroutine()

	for testNum := 1; testNum <= testCount; testNum++ {
		func() {
			t.Logf("Running test #%d\n", testNum)

			serverCfg := server.NewConfig(net.IPv4zero.String(), 8375, useConnectionless)
			echoServer, err := server.New(serverCfg)
			if err != nil {
				t.Error(err)
				return
			}

			go func() {
				for conn := range echoServer.Accept() {
					go func(c server.Connection) {
						buffer := make([]byte, serverCfg.ReadBufferSize)
						for {
							count, e := c.Read(buffer)
							if e != nil {
								if errors.Is(e, io.EOF) {
									return
								}

								t.Errorf("connection read error: %v", e)
								return
							}

							if count < 1 {
								continue
							}

							_, e = c.Write(buffer[:count])
							if e != nil {
								if errors.Is(err, io.EOF) {
									return
								}

								t.Errorf("connection write error: %v", e)
								return
							}
						}
					}(conn)
				}
			}()

			err = echoServer.Start()
			if err != nil {
				t.Error(err)
				return
			}
			defer echoServer.Stop()

			go func() {
				for se := range echoServer.Errors() {
					t.Logf("server error: %v\n", se)
				}
			}()

			time.Sleep(time.Second)

			var testPassCount atomic.Int32
			var wg sync.WaitGroup
			wg.Add(clientCount)
			for i := int32(0); i < clientCount; i++ {
				idx := i
				go func() {
					defer wg.Done()

					clientCfg := client.NewConfig(net.IPv4zero.String(), 8375, useConnectionless)
					c, e := client.New(clientCfg)
					if e != nil {
						t.Error(e)
						return
					}

					e = c.Start()
					if e != nil {
						t.Error(e)
						return
					}
					defer func() {
						de := c.Stop()
						if de != nil {
							t.Error(de)
						}
					}()

					request := GetRandomString(requestLength)
					count, e := c.Write(request)
					if e != nil {
						t.Error(e)
						return
					}
					if count != len(request) {
						t.Error(fmt.Errorf("client #%d error: expected to write %d bytes, wrote %d instead", idx, len(request), count))
						return
					}

					time.Sleep(time.Second)

					response := make([]byte, len(request))
					readCount, e := c.Read(response)
					if e != nil {
						t.Error(e)
						return
					}
					if readCount != len(request) {
						t.Error(fmt.Errorf("client #%d error: expected to read %d bytes, read %d instead", idx, len(request), readCount))
						return
					}
					if string(response) != string(request) {
						t.Error(fmt.Errorf("client #%d error: expected to receive '%s', received '%s' instead", idx, string(request), string(response)))
						return
					}

					testPassCount.Add(1)
				}()
			}

			wg.Wait()

			if testPassCount.Load() != clientCount {
				t.Errorf("test pass count is %d, expected %d", testPassCount.Load(), clientCount)
				return
			}
		}()
	}

	time.Sleep(2 * time.Second)

	finishingRoutineCount := runtime.NumGoroutine()
	if finishingRoutineCount > startingRoutineCount {
		t.Errorf("unexpected thread count (expected <=%d, have %d)", startingRoutineCount, finishingRoutineCount)
	}
}
