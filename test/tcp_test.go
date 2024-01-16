package test

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
	"tonysoft.com/comm/pkg/client"
	"tonysoft.com/comm/pkg/server"
	"tonysoft.com/comm/pkg/stream"
)

func TestTcpClientReadWrite(t *testing.T) {
	ips, err := net.LookupIP("google.com")
	if err != nil {
		t.Error(err)
		return
	}
	ip := ips[0]

	cfg := client.NewConfig(ip.String(), 80)
	c, err := client.New(cfg)
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

		if c.IsConnected() {
			t.Error("expected the client to return 'false' from IsConnected()")
			return
		}
	}()

	if !c.IsConnected() {
		t.Error("expected the client to return 'true' from IsConnected()")
		return
	}

	request := []byte("GET / HTTP/1.1\r\nHost: www.google.com\r\n\r\n")
	count, err := c.Write(request)
	if err != nil {
		t.Error(err)
		return
	}
	if count != len(request) {
		t.Error(fmt.Errorf("expected to write %d bytes, wrote %d instead", len(request), count))
		return
	}

	time.Sleep(3 * time.Second)

	response := "HTTP/1.1 200 OK"
	buffer := make([]byte, len(response))
	count, err = c.Read(buffer)
	if err != nil {
		t.Error(err)
		return
	}
	if count != len(response) {
		t.Error(fmt.Errorf("expected to read %d bytes, read %d instead", len(response), count))
		return
	}
	if string(buffer) != response {
		t.Error(fmt.Errorf("expected to receive '%s', received '%s' instead", response, string(buffer)))
		return
	}
}

func TestTcpServerReadWrite(t *testing.T) {
	testPassed := false
	testPing := []byte("ping")
	testPong := []byte("pong")

	serverCfg := server.NewConfig(net.IPv4zero.String(), 8375)
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

	clientCfg := client.NewConfig(net.IPv4zero.String(), 8375)
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

func TestTcpServerMultiClient(t *testing.T) {
	const testCount = 5
	const clientCount = 50
	const requestLength = 500

	startingRoutineCount := runtime.NumGoroutine()

	for testNum := 1; testNum <= testCount; testNum++ {
		func() {
			t.Logf("Running test #%d\n", testNum)

			serverCfg := server.NewConfig(net.IPv4zero.String(), 8375)
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
								if errors.Is(e, syscall.ECONNRESET) || errors.Is(e, net.ErrClosed) {
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
								if errors.Is(err, syscall.ECONNRESET) {
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

					clientCfg := client.NewConfig(net.IPv4zero.String(), 8375)
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

func TestTcpServerMultiClientUsingStringStream(t *testing.T) {
	const testCount = 5
	const clientCount = 50
	const requestLength = 500

	startingRoutineCount := runtime.NumGoroutine()

	for testNum := 1; testNum <= testCount; testNum++ {
		func() {
			t.Logf("Running test #%d\n", testNum)

			serverCfg := server.NewConfig(net.IPv4zero.String(), 8375)
			echoServer, err := server.New(serverCfg)
			if err != nil {
				t.Error(err)
				return
			}

			go func() {
				for conn := range echoServer.Accept() {
					go func(c server.Connection) {
						// With error checking:
						ss := &stream.StringStream{}
						for str := range ss.Stream(c) {
							_, e := c.Write([]byte(str))
							if e != nil {
								if errors.Is(e, syscall.ECONNRESET) {
									return
								}

								t.Errorf("connection write error: %v", e)
								return
							}
						}
						if ss.Error() != nil {
							if errors.Is(ss.Error(), syscall.ECONNRESET) || errors.Is(ss.Error(), net.ErrClosed) {
								return
							}

							t.Errorf("connection read error: %v", ss.Error())
							return
						}

						// With no error checking:
						//for str := range stream.String(conn) {
						//	conn.Write([]byte(str))
						//}
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

					clientCfg := client.NewConfig(net.IPv4zero.String(), 8375)
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

					request := GetRandomCString(requestLength)
					count, e := c.Write(request)
					if e != nil {
						t.Error(e)
						return
					}
					if count != len(request) {
						t.Error(fmt.Errorf("client #%d error: expected to write %d bytes, wrote %d instead", idx, len(request), count))
						return
					}

					request = request[:len(request)-1]
					time.Sleep(2 * time.Second)

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
